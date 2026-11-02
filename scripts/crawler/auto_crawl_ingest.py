#!/usr/bin/env python3
"""
Automated Recipe Crawler & AI Ingestion Pipeline for what-to-eat
Fetches recipes from Dien May Xanh, structures them using Gemini AI,
resolves and auto-creates ingredients, and ingests them into MongoDB.
"""

import os
import re
import sys
import time
import unicodedata
from datetime import datetime, timezone
from typing import List, Optional

import requests
from bs4 import BeautifulSoup
from dotenv import load_dotenv
from pydantic import BaseModel, Field
from pymongo import MongoClient
from bson import ObjectId

# Load environment variables from .env if present
load_dotenv()

# ==============================================================================
# Configuration
# ==============================================================================
MONGODB_URI = os.getenv("MONGODB_CONNECTION_STRING")
DATABASE_NAME = os.getenv("DATABASE_NAME", "what-to-eat")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")

DEFAULT_CATEGORIES = ["mon-chien", "mon-xao", "mon-kho", "mon-canh", "mon-nuong", "mon-hap"]
CRAWL_CATEGORIES = [
    c.strip() for c in os.getenv("CRAWL_CATEGORIES", ",".join(DEFAULT_CATEGORIES)).split(",") if c.strip()
]
DAILY_CRAWL_LIMIT = int(os.getenv("DAILY_CRAWL_LIMIT", "5"))

if not GEMINI_API_KEY:
    print("[ERROR] GEMINI_API_KEY is not set. Please set it in your environment or secrets.", file=sys.stderr)
    sys.exit(1)

if not MONGODB_URI:
    print("[ERROR] MONGODB_CONNECTION_STRING is not set. Please set it in your environment or secrets.", file=sys.stderr)
    sys.exit(1)

# Initialize clients
from google import genai
from google.genai import types

ai_client = genai.Client(api_key=GEMINI_API_KEY)
mongo_client = MongoClient(MONGODB_URI)
db = mongo_client[DATABASE_NAME]
dishes_col = db["Dishes"]
ingredients_col = db["Ingredients"]

# ==============================================================================
# Pydantic Extraction Schemas
# ==============================================================================
class ParsedIngredient(BaseModel):
    name_vi: str = Field(description="Tên nguyên liệu tiếng Việt chuẩn, ví dụ: 'Thịt ba chỉ'")
    name_en: str = Field(description="Standard English ingredient name, e.g. 'Pork belly'")
    quantity: float = Field(default=1.0, description="Số lượng dạng số thực (ví dụ 300 cho 300g, 2 cho 2 quả)")
    measure: str = Field(default="g", description="Đơn vị tính, ví dụ: 'g', 'kg', 'ml', 'quả', 'muỗng canh', 'nhánh'")
    note: Optional[str] = Field(default=None, description="Ghi chú thêm về nguyên liệu nếu có")
    category: str = Field(
        default="OTHER",
        description="Must be one of: 'PROTEIN FOODS', 'VEGETABLES', 'GRAINS', 'FATS AND OILS', 'CONDIMENTS AND SAUCES', 'HERBS AND SPICES', 'MILK AND DAIRY', 'FRUIT', 'NUTS AND SEEDS', 'OTHER'"
    )

class RecipeExtraction(BaseModel):
    title_vi: str = Field(description="Tên món ăn tiếng Việt ngắn gọn, hấp dẫn")
    title_en: str = Field(description="English translation of the dish title")
    short_desc_vi: str = Field(description="Mô tả ngắn gọn về hương vị món ăn (1-2 câu tiếng Việt)")
    short_desc_en: str = Field(description="Short description in English (1-2 sentences)")
    content_vi: str = Field(description="Các bước thực hiện chi tiết bằng tiếng Việt, định dạng HTML ngữ nghĩa (<p>, <h4>, <ul>, <li>)")
    content_en: str = Field(description="Detailed step-by-step cooking instructions in English formatted in clean semantic HTML")
    thumbnail: Optional[str] = Field(default=None, description="Direct URL to high-resolution image of finished dish")
    preparation_time: int = Field(default=15, description="Thời gian chuẩn bị ước tính (phút)")
    cooking_time: int = Field(default=20, description="Thời gian nấu ước tính (phút)")
    difficulty_level: str = Field(default="MEDIUM", description="Độ khó: 'EASY', 'MEDIUM', or 'HARD'")
    meal_categories: List[str] = Field(
        default=["LUNCH", "DINNER"],
        description="List of applicable categories from: ['BREAKFAST', 'LUNCH', 'DINNER', 'SNACK', 'FRIED', 'GRILLED', 'STEAMED', 'BAKED', 'SOUP', 'NOODLE', 'RICE', 'SEAFOOD', 'MEAT', 'VEGETARIAN', 'TRADITIONAL', 'STREET_FOOD']"
    )
    tags: List[str] = Field(default=[], description="Từ khóa tìm kiếm không dấu và có dấu")
    ingredients: List[ParsedIngredient]

# ==============================================================================
# Helper Functions
# ==============================================================================
def slugify(text: str) -> str:
    """Converts Vietnamese string into clean URL slug."""
    text = unicodedata.normalize('NFKD', text)
    # replace Vietnamese characters like đ, Đ
    text = text.replace('đ', 'd').replace('Đ', 'D')
    text = text.encode('ascii', 'ignore').decode('utf-8')
    text = re.sub(r'[^\w\s-]', '', text).strip().lower()
    return re.sub(r'[-\s]+', '-', text)

def fetch_category_recipe_links(category: str, limit_per_cat: int = 10) -> List[str]:
    """Fetches list of recipe links from a category page."""
    base_url = "https://www.dienmayxanh.com/vao-bep"
    cat_url = f"{base_url}/{category}" if not category.startswith("http") else category
    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
    }
    
    try:
        resp = requests.get(cat_url, headers=headers, timeout=15)
        resp.raise_for_status()
    except Exception as e:
        print(f"  [WARN] Failed to load category {cat_url}: {e}")
        return []

    soup = BeautifulSoup(resp.text, "html.parser")
    links = []
    
    for a in soup.select("ul.listdish li a, .cate-cook li a, div.news-list a"):
        href = a.get("href")
        if href and "/vao-bep/" in href and not href.endswith(f"/vao-bep/{category}"):
            if not href.startswith("http"):
                href = f"https://www.dienmayxanh.com{href}"
            links.append(href)
            
    return list(dict.fromkeys(links))[:limit_per_cat]

def parse_recipe_with_ai(page_html: str, url: str) -> RecipeExtraction:
    """Uses Gemini 2.5 Flash to extract structured recipe data."""
    soup = BeautifulSoup(page_html, "html.parser")

    # Extract og:image fallback
    og_img = soup.find("meta", property="og:image")
    meta_img_url = og_img["content"] if og_img and og_img.get("content") else None

    # Fallback to first article image
    first_img = None
    if not meta_img_url:
        img_tag = soup.select_one("article img, .content-cook img, div.cook-content img")
        if img_tag and img_tag.get("src"):
            first_img = img_tag["src"]

    fallback_img = meta_img_url or first_img

    # Remove non-content elements to conserve token context
    for el in soup(["script", "style", "nav", "footer", "header", "noscript", "aside"]):
        el.extract()

    # Get main text
    body_text = soup.get_text(separator="\n", strip=True)

    prompt = f"""
    You are an expert culinary assistant and database architect.
    Extract the recipe from this Vietnamese cooking page: {url}
    Fallback recipe image candidate: {fallback_img}

    Instructions:
    1. Extract a clear title in Vietnamese and translate it naturally into English.
    2. Write a short, appetizing description (1-2 sentences) in both Vietnamese and English.
    3. Format cooking steps clearly into semantic HTML (<p>, <h4>, <ul>, <li>) for both Vietnamese and English.
    4. Detect realistic preparation and cooking times in minutes.
    5. Determine difficulty: 'EASY', 'MEDIUM', or 'HARD'.
    6. Select applicable meal categories (e.g. FRIED, LUNCH, DINNER, TRADITIONAL, MEAT, SEAFOOD, etc.).
    7. Extract all ingredients precisely with numeric quantity, unit/measure, notes, and nutritional category.
    8. Set thumbnail to the best recipe image URL found.

    Page Text:
    {body_text[:14000]}
    """

    response = ai_client.models.generate_content(
        model="gemini-2.5-flash",
        contents=prompt,
        config=types.GenerateContentConfig(
            response_mime_type="application/json",
            response_schema=RecipeExtraction,
            temperature=0.2,
        ),
    )
    return RecipeExtraction.model_validate_json(response.text)

def get_or_create_ingredient(ing: ParsedIngredient, author_id: str = "system-crawler") -> str:
    """Finds existing ingredient by slug or auto-creates it in Ingredients collection."""
    slug = slugify(ing.name_vi)
    if not slug:
        slug = slugify(ing.name_en)

    now = datetime.now(timezone.utc)

    # Search for existing ingredient by slug
    existing = ingredients_col.find_one({"slug": slug, "deleted": False})
    if existing:
        return str(existing["_id"])

    # Also check case-insensitive Vietnamese title match
    existing_by_title = ingredients_col.find_one({
        "title.data": {"$regex": f"^{re.escape(ing.name_vi)}$", "$options": "i"},
        "deleted": False
    })
    if existing_by_title:
        return str(existing_by_title["_id"])

    # Auto-create new Ingredient document
    new_doc = {
        "_id": ObjectId(),
        "slug": slug,
        "title": [
            {"lang": "vi", "data": ing.name_vi},
            {"lang": "en", "data": ing.name_en}
        ],
        "measure": ing.measure,
        "ingredientCategory": [ing.category],
        "images": [],
        "deleted": False,
        "createdAt": now,
        "updatedAt": now,
        "createdBy": author_id,
        "updatedBy": author_id,
    }
    ingredients_col.insert_one(new_doc)
    print(f"  [+] Auto-created new Ingredient: {ing.name_vi} (slug: {slug}, id: {new_doc['_id']})")
    return str(new_doc["_id"])

def ingest_recipe(url: str, author_id: str = "github-actions-crawler") -> bool:
    """Scrapes, parses with AI, maps ingredients, and saves dish to MongoDB."""
    url_slug = slugify(url.rstrip("/").split("/")[-1])
    
    # Check if dish already exists in DB
    if dishes_col.find_one({"slug": url_slug, "deleted": False}):
        print(f"[-] Dish already exists with slug '{url_slug}'. Skipping.")
        return False

    print(f"\n[>] Processing recipe: {url}")
    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
    }
    resp = requests.get(url, headers=headers, timeout=20)
    resp.raise_for_status()

    # AI structuring
    extracted = parse_recipe_with_ai(resp.text, url)
    print(f"  [AI] Extracted: {extracted.title_vi} | {extracted.title_en}")

    # Use generated title slug if cleaner than url slug
    dish_slug = slugify(extracted.title_vi) or url_slug
    
    # Check if slug already in DB
    if dishes_col.find_one({"slug": dish_slug, "deleted": False}):
        dish_slug = f"{dish_slug}-{int(time.time())}"

    # Resolve / Auto-create ingredients
    mapped_ingredients = []
    for ing in extracted.ingredients:
        ing_id = get_or_create_ingredient(ing, author_id=author_id)
        mapped_ingredients.append({
            "ingredientId": ing_id,
            "slug": slugify(ing.name_vi),
            "quantity": ing.quantity,
            "note": ing.note
        })

    now = datetime.now(timezone.utc)
    
    # Create Dish document matching model.Dish
    dish_doc = {
        "_id": ObjectId(),
        "slug": dish_slug,
        "title": [
            {"lang": "vi", "data": extracted.title_vi},
            {"lang": "en", "data": extracted.title_en}
        ],
        "shortDescription": [
            {"lang": "vi", "data": extracted.short_desc_vi},
            {"lang": "en", "data": extracted.short_desc_en}
        ],
        "content": [
            {"lang": "vi", "data": extracted.content_vi},
            {"lang": "en", "data": extracted.content_en}
        ],
        "tags": [t.strip().lower() for t in extracted.tags if t.strip()],
        "preparationTime": float(extracted.preparation_time),
        "cookingTime": float(extracted.cooking_time),
        "difficultLevel": extracted.difficulty_level,
        "mealCategories": extracted.meal_categories,
        "ingredientCategories": list(set(ing.category for ing in extracted.ingredients)),
        "thumbnail": extracted.thumbnail,
        "videos": [],
        "ingredients": mapped_ingredients,
        "relatedDishes": [],
        "labels": ["AUTO_CRAWLED", "DIENMAYXANH"],
        "deleted": False,
        "createdAt": now,
        "updatedAt": now,
        "createdBy": author_id,
        "updatedBy": author_id
    }

    dishes_col.insert_one(dish_doc)
    print(f"[SUCCESS] Successfully ingested dish '{extracted.title_vi}' (ID: {dish_doc['_id']}, Slug: {dish_slug}) with {len(mapped_ingredients)} ingredients.")
    return True

# ==============================================================================
# Main Runner
# ==============================================================================
def main():
    print("=" * 60)
    print("Starting What-To-Eat Recipe Ingestion Job")
    print(f"Categories to scan: {CRAWL_CATEGORIES}")
    print(f"Daily crawl target limit: {DAILY_CRAWL_LIMIT}")
    print("=" * 60)

    total_inserted = 0

    for cat in CRAWL_CATEGORIES:
        if total_inserted >= DAILY_CRAWL_LIMIT:
            break

        print(f"\nScanning category '{cat}'...")
        links = fetch_category_recipe_links(cat, limit_per_cat=10)
        print(f"Found {len(links)} candidate links.")

        for link in links:
            if total_inserted >= DAILY_CRAWL_LIMIT:
                print(f"\nReached daily target limit of {DAILY_CRAWL_LIMIT} dishes. Finishing job.")
                break

            try:
                success = ingest_recipe(link)
                if success:
                    total_inserted += 1
                    time.sleep(2)  # Polite crawl delay between requests
            except Exception as err:
                print(f"[ERROR] Failed to process {link}: {err}", file=sys.stderr)

    print("\n" + "=" * 60)
    print(f"Job completed. Ingested {total_inserted} new dishes.")
    print("=" * 60)

if __name__ == "__main__":
    main()

