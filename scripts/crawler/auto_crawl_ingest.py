#!/usr/bin/env python3
"""
Automated Recipe Ingestion Pipeline for what-to-eat
Uses Google Gemini 2.5 Flash with Google Search Grounding to discover
authentic Vietnamese recipes from the web (including Dien May Xanh, Cookpad, etc.),
structures them into bilingual format, records the original source URL,
resolves and auto-creates ingredients, and ingests them into MongoDB.
"""

import os
import re
import sys
import time
import unicodedata
from datetime import datetime, timezone
from typing import List, Optional

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
GEMINI_MODEL = os.getenv("GEMINI_MODEL", "gemini-3.8-flash")

DEFAULT_CATEGORIES = [
    "món chiên giòn ngon",
    "món xào đưa cơm",
    "món kho đậm đà",
    "món canh thanh mát",
    "món nướng thơm lừng",
    "món hấp thanh đạm"
]

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
# Pydantic Schemas for AI Structured Extraction
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
    source_url: str = Field(description="Original website URL where the recipe was found (e.g. from dienmayxanh.com/vao-bep, cookpad.com/vn, monngonmoingay.com, etc.)")
    preparation_time: int = Field(default=15, description="Thời gian chuẩn bị ước tính (phút)")
    cooking_time: int = Field(default=20, description="Thời gian nấu ước tính (phút)")
    difficulty_level: str = Field(default="MEDIUM", description="Độ khó: 'EASY', 'MEDIUM', or 'HARD'")
    meal_categories: List[str] = Field(
        default=["LUNCH", "DINNER"],
        description="List of applicable categories from: ['BREAKFAST', 'LUNCH', 'DINNER', 'SNACK', 'FRIED', 'GRILLED', 'STEAMED', 'BAKED', 'SOUP', 'NOODLE', 'RICE', 'SEAFOOD', 'MEAT', 'VEGETARIAN', 'TRADITIONAL', 'STREET_FOOD']"
    )
    tags: List[str] = Field(default=[], description="Từ khóa tìm kiếm không dấu và có dấu")
    ingredients: List[ParsedIngredient]

class RecipeBatch(BaseModel):
    recipes: List[RecipeExtraction]

# ==============================================================================
# Helper Functions
# ==============================================================================
def slugify(text: str) -> str:
    """Converts Vietnamese string into clean URL slug."""
    text = unicodedata.normalize('NFKD', text)
    text = text.replace('đ', 'd').replace('Đ', 'D')
    text = text.encode('ascii', 'ignore').decode('utf-8')
    text = re.sub(r'[^\w\s-]', '', text).strip().lower()
    return re.sub(r'[-\s]+', '-', text)

def discover_and_structure_recipes(category_keyword: str, count: int = 2) -> List[RecipeExtraction]:
    """
    Uses Gemini 2.5 Flash with Google Search Grounding to find authentic
    Vietnamese recipes on the web and structure them into our schema.
    """
    print(f"\n[AI Search] Searching Google for '{category_keyword}' recipes...")
    
    search_prompt = f"""
    Search Google for authentic, popular Vietnamese cooking recipes matching: '{category_keyword}'.
    Look up real recipe pages from reputable Vietnamese culinary websites (such as dienmayxanh.com/vao-bep, cookpad.com/vn, monngonmoingay.com, vnexpress.net/du-lich/am-thuc, bepgiadinh.com).
    
    Select {count} distinct popular recipes that Vietnamese families love to eat.
    For each recipe, extract:
    1. Vietnamese title and English title.
    2. The exact original website URL (source_url) where this recipe was published.
    3. Short description in VI and EN.
    4. Clear step-by-step instructions in VI and EN.
    5. Direct high-res image thumbnail URL.
    6. Realistic preparation time and cooking time in minutes.
    7. Difficulty level (EASY, MEDIUM, HARD).
    8. Meal categories (FRIED, SOUP, MEAT, SEAFOOD, etc.) and tags.
    9. Complete ingredient list with amounts, units, and notes.
    """

    # Step 1: Search the web via Google Search Grounding
    search_response = ai_client.models.generate_content(
        model=GEMINI_MODEL,
        contents=search_prompt,
        config=types.GenerateContentConfig(
            tools=[types.Tool(google_search=types.GoogleSearch())],
            temperature=0.3,
        ),
    )

    # Extract source URLs from grounding metadata chunks if available
    grounding_uris = []
    if search_response.candidates and search_response.candidates[0].grounding_metadata:
        chunks = search_response.candidates[0].grounding_metadata.grounding_chunks or []
        for c in chunks:
            if hasattr(c, "web") and c.web and c.web.uri:
                grounding_uris.append(c.web.uri)

    grounding_hint = f"\nRelevant source web URLs discovered: {', '.join(grounding_uris[:8])}" if grounding_uris else ""

    # Step 2: Enforce strict structured JSON schema output
    format_prompt = f"""
    Convert the following grounded recipe search results into the required JSON schema.
    IMPORTANT: Ensure `source_url` accurately contains the original web URL where the recipe is found.
    {grounding_hint}

    Search Results:
    {search_response.text}
    """

    structured_response = ai_client.models.generate_content(
        model=GEMINI_MODEL,
        contents=format_prompt,
        config=types.GenerateContentConfig(
            response_mime_type="application/json",
            response_schema=RecipeBatch,
            temperature=0.2,
        ),
    )

    batch = RecipeBatch.model_validate_json(structured_response.text)
    return batch.recipes

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

    # Check case-insensitive Vietnamese title match
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
    print(f"    [+] Auto-created new Ingredient: {ing.name_vi} (slug: {slug})")
    return str(new_doc["_id"])

def ingest_recipe(recipe: RecipeExtraction, author_id: str = "github-actions-crawler") -> bool:
    """Resolves ingredients and inserts the recipe document with sourceUrl into MongoDB."""
    dish_slug = slugify(recipe.title_vi)
    if not dish_slug:
        dish_slug = slugify(recipe.title_en)

    # Check if dish already exists by slug or source_url
    existing_query = {
        "$or": [
            {"slug": dish_slug},
            {"sourceUrl": recipe.source_url}
        ],
        "deleted": False
    }
    if dishes_col.find_one(existing_query):
        print(f"  [-] Dish already exists (slug: '{dish_slug}' or sourceUrl: '{recipe.source_url}'). Skipping.")
        return False

    print(f"\n  [>] Ingesting Dish: {recipe.title_vi} ({recipe.title_en})")
    print(f"      Source URL: {recipe.source_url}")

    # Resolve / Auto-create ingredients
    mapped_ingredients = []
    for ing in recipe.ingredients:
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
            {"lang": "vi", "data": recipe.title_vi},
            {"lang": "en", "data": recipe.title_en}
        ],
        "shortDescription": [
            {"lang": "vi", "data": recipe.short_desc_vi},
            {"lang": "en", "data": recipe.short_desc_en}
        ],
        "content": [
            {"lang": "vi", "data": recipe.content_vi},
            {"lang": "en", "data": recipe.content_en}
        ],
        "tags": [t.strip().lower() for t in recipe.tags if t.strip()],
        "preparationTime": float(recipe.preparation_time),
        "cookingTime": float(recipe.cooking_time),
        "difficultLevel": recipe.difficulty_level,
        "mealCategories": recipe.meal_categories,
        "ingredientCategories": list(set(ing.category for ing in recipe.ingredients)),
        "thumbnail": recipe.thumbnail,
        "videos": [],
        "ingredients": mapped_ingredients,
        "relatedDishes": [],
        "labels": ["AUTO_CRAWLED", "GEMINI_SEARCH"],
        "sourceUrl": recipe.source_url,
        "deleted": False,
        "createdAt": now,
        "updatedAt": now,
        "createdBy": author_id,
        "updatedBy": author_id
    }

    dishes_col.insert_one(dish_doc)
    print(f"  [SUCCESS] Inserted '{recipe.title_vi}' (ID: {dish_doc['_id']}, Slug: {dish_slug}) with {len(mapped_ingredients)} ingredients.\n")
    return True

# ==============================================================================
# Main Runner
# ==============================================================================
def main():
    print("=" * 60)
    print("Starting What-To-Eat AI Search & Ingestion Job")
    print(f"Categories to explore: {CRAWL_CATEGORIES}")
    print(f"Daily crawl target limit: {DAILY_CRAWL_LIMIT}")
    print("=" * 60)

    total_inserted = 0
    error_count = 0

    for category in CRAWL_CATEGORIES:
        if total_inserted >= DAILY_CRAWL_LIMIT:
            break

        remaining = DAILY_CRAWL_LIMIT - total_inserted
        fetch_count = min(remaining, 2)

        try:
            recipes = discover_and_structure_recipes(category, count=fetch_count)
            print(f"  Found {len(recipes)} structured recipes for '{category}'.")

            for recipe in recipes:
                if total_inserted >= DAILY_CRAWL_LIMIT:
                    break
                success = ingest_recipe(recipe)
                if success:
                    total_inserted += 1
                time.sleep(1)

        except Exception as err:
            print(f"[ERROR] Failed during category '{category}': {err}")
            error_count += 1

    print("=" * 60)
    print(f"Job completed. Ingested {total_inserted} new dishes. Total category errors: {error_count}")
    print("=" * 60)

    if total_inserted == 0 and error_count > 0:
        print("[FAIL] No dishes were ingested due to API or network errors.", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
