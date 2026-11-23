#!/usr/bin/env python3
"""
Resilient World Recipe Ingestion Pipeline for what-to-eat
Multi-tiered ingestion pipeline that NEVER stops and guarantees dish ingestion from cuisines all over the world:
  - Tier 1: Gemini 3.8 Flash with Google Search Grounding
  - Tier 2: Gemini Direct AI Recipe Generation (no search quota needed, bypasses 429)
  - Tier 3: Curated World Recipe Pool (guarantees ingestion even if AI quota is 100% exhausted)
"""

import json
import math
import os
import random
import re
import sys
import time
import unicodedata
import urllib.parse
from datetime import datetime, timezone
from typing import List, Optional, Tuple

import requests
from bs4 import BeautifulSoup
from dotenv import load_dotenv
from pydantic import BaseModel, Field
from pymongo import MongoClient
from bson import ObjectId

# Load environment variables
load_dotenv()

# ==============================================================================
# Configuration
# ==============================================================================
MONGODB_URI = os.getenv("MONGODB_CONNECTION_STRING")
DATABASE_NAME = os.getenv("DATABASE_NAME", "what-to-eat")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")
GEMINI_MODEL = os.getenv("GEMINI_MODEL", "gemini-3.8-flash")

DEFAULT_CATEGORIES = [
    "Italian cuisine",
    "Japanese cuisine",
    "Mexican cuisine",
    "Indian cuisine",
    "Thai cuisine",
    "Korean cuisine",
    "French cuisine",
    "Chinese cuisine",
    "Greek and Mediterranean cuisine",
    "Middle Eastern cuisine",
    "Spanish cuisine",
    "American comfort food",
    "Vietnamese cuisine",
    "Moroccan and North African cuisine",
    "Turkish cuisine",
    "Latin American cuisine",
]

_raw_categories = os.getenv("CRAWL_CATEGORIES", "").strip()
CRAWL_CATEGORIES = [
    c.strip() for c in _raw_categories.split(",") if c.strip()
] or list(DEFAULT_CATEGORIES)
DAILY_CRAWL_LIMIT = int(os.getenv("DAILY_CRAWL_LIMIT", "2"))

if not MONGODB_URI:
    print("[ERROR] MONGODB_CONNECTION_STRING is not set. Please configure it.", file=sys.stderr)
    sys.exit(1)

# Initialize MongoDB client
mongo_client = MongoClient(MONGODB_URI)
db = mongo_client[DATABASE_NAME]
dishes_col = db["Dishes"]
ingredients_col = db["Ingredients"]

# Initialize Gemini client if key is available
ai_client = None
if GEMINI_API_KEY:
    try:
        from google import genai
        from google.genai import types
        ai_client = genai.Client(api_key=GEMINI_API_KEY)
    except Exception as e:
        print(f"[WARN] Failed to initialize Gemini Client: {e}")

# ==============================================================================
# Pydantic Schemas for AI Structured Extraction
# ==============================================================================
class ParsedIngredient(BaseModel):
    name_vi: str = Field(description="Tên nguyên liệu tiếng Việt chuẩn, ví dụ: 'Phô mai Parmesan', 'Dầu ô liu'")
    name_en: str = Field(description="Standard English ingredient name, e.g. 'Parmesan cheese', 'Olive oil'")
    quantity: float = Field(default=1.0, description="Numeric quantity (e.g. 300 for 300g, 2 for 2 pieces)")
    measure: str = Field(default="g", description="Unit of measurement, e.g. 'g', 'kg', 'ml', 'tbsp', 'tsp', 'cup', 'piece', 'clove'")
    note: Optional[str] = Field(default=None, description="Additional preparation note if any (e.g. 'finely chopped', 'grated')")
    category: str = Field(
        default="OTHER",
        description="Must be one of: 'PROTEIN FOODS', 'VEGETABLES', 'GRAINS', 'FATS AND OILS', 'CONDIMENTS AND SAUCES', 'HERBS AND SPICES', 'MILK AND DAIRY', 'FRUIT', 'NUTS AND SEEDS', 'OTHER'"
    )

class RecipeExtraction(BaseModel):
    title_vi: str = Field(description="Tên món ăn tiếng Việt ngắn gọn, hấp dẫn (ví dụ: 'Mì Ý Carbonara')")
    title_en: str = Field(description="Authentic dish title in English (e.g. 'Spaghetti Carbonara')")
    short_desc_vi: str = Field(description="Mô tả ngắn gọn về hương vị và xuất xứ món ăn (1-2 câu tiếng Việt)")
    short_desc_en: str = Field(description="Short description of the dish's flavor and culinary origin in English (1-2 sentences)")
    content_vi: str = Field(description="Các bước thực hiện chi tiết bằng tiếng Việt, định dạng HTML ngữ nghĩa (<p>, <h4>, <ul>, <li>)")
    content_en: str = Field(description="Detailed step-by-step cooking instructions in English formatted in clean semantic HTML (<p>, <h4>, <ul>, <li>)")
    thumbnail: Optional[str] = Field(default=None, description="Direct URL to high-resolution image of finished dish")
    source_url: str = Field(description="Original website URL where the recipe was found (e.g. seriouseats.com, bbcgoodfood.com, allrecipes.com, recipetineats.com, justonecookbook.com, bonappetit.com, foodnetwork.com, cookpad.com, etc.)")
    preparation_time: int = Field(default=15, description="Estimated preparation time in minutes")
    cooking_time: int = Field(default=20, description="Estimated cooking time in minutes")
    difficulty_level: str = Field(default="MEDIUM", description="Difficulty level: 'EASY', 'MEDIUM', or 'HARD'")
    meal_categories: List[str] = Field(
        default=["LUNCH", "DINNER"],
        description="List of applicable categories from: ['BREAKFAST', 'LUNCH', 'DINNER', 'SNACK', 'FRIED', 'GRILLED', 'STEAMED', 'BAKED', 'SOUP', 'NOODLE', 'RICE', 'SEAFOOD', 'MEAT', 'VEGETARIAN', 'TRADITIONAL', 'STREET_FOOD']"
    )
    tags: List[str] = Field(default=[], description="Bilingual search keywords in English and Vietnamese including cuisine/country of origin (e.g. 'italian', 'món ý', 'pasta', 'carbonara')")
    ingredients: List[ParsedIngredient]

class RecipeBatch(BaseModel):
    recipes: List[RecipeExtraction]

# ==============================================================================
# Helpers
# ==============================================================================
STOPWORDS = {"with", "and", "the", "for", "mon", "cach", "style", "homemade", "authentic"}

BROWSER_UA = (
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
    "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
)
WIKI_UA = "WhatToEatRecipeCrawler/1.0 (https://github.com/fxanhkhoa/what-to-eat-api)"

REJECTED_IMAGE_PATTERNS = (
    ".svg", ".ico", ".gif", "logo", "favicon", "sprite", "placeholder",
    "avatar", "1x1", "pixel", "flag_of", "coat_of_arms", "locator",
    "diagram", "emblem", "map_of",
)

LEADING_CUISINE_PREFIXES = (
    "middle eastern", "north african", "latin american",
    "traditional", "authentic", "classic", "homemade", "easy",
    "best", "spicy", "crispy", "quick", "italian", "japanese",
    "mexican", "indian", "thai", "korean", "french", "chinese",
    "greek", "mediterranean", "spanish", "american", "vietnamese",
    "moroccan", "turkish", "brazilian", "sichuan",
)

def _headers_for_url(url: str) -> dict:
    """Selects an appropriate User-Agent header for Wikimedia vs general websites."""
    host = urllib.parse.urlsplit(url).netloc.lower()
    ua = WIKI_UA if ("wikimedia.org" in host or "wikipedia.org" in host) else BROWSER_UA
    return {"User-Agent": ua}

def _clean_tracking_params(url: str) -> str:
    """Strips utm_* query parameters from a URL while preserving functional query args."""
    if not url:
        return url
    parts = urllib.parse.urlsplit(url)
    if not parts.query:
        return url
    filtered_qs = [
        (k, v) for k, v in urllib.parse.parse_qsl(parts.query, keep_blank_values=True)
        if not k.lower().startswith("utm_")
    ]
    return urllib.parse.urlunsplit(
        (parts.scheme, parts.netloc, parts.path, urllib.parse.urlencode(filtered_qs), parts.fragment)
    )

def _validate_image_response(resp) -> bool:
    """Checks HTTP status, Content-Type, and Content-Length of an image response."""
    if resp.status_code != 200:
        return False
    content_type = (resp.headers.get("Content-Type") or "").split(";")[0].strip().lower()
    if not content_type.startswith("image/") or "svg" in content_type or "icon" in content_type:
        return False
    content_length = resp.headers.get("Content-Length")
    if content_length is not None:
        try:
            if int(content_length) < 5000:
                return False
        except (ValueError, TypeError):
            pass
    return True

def is_valid_image_url(url: Optional[str]) -> bool:
    """Verifies that a candidate URL resolves to a real, sufficiently large raster image."""
    if not url or not isinstance(url, str):
        return False
    cleaned = url.strip()
    if not cleaned.startswith(("http://", "https://")):
        return False

    lower_path = urllib.parse.urlsplit(cleaned).path.lower()
    lower_full = cleaned.lower()
    if any(pat in lower_path or pat in lower_full for pat in REJECTED_IMAGE_PATTERNS):
        return False

    headers = _headers_for_url(cleaned)
    try:
        resp = requests.head(cleaned, headers=headers, allow_redirects=True, timeout=6)
        if resp.status_code in (400, 403, 405):
            resp.close()
            with requests.get(cleaned, headers=headers, allow_redirects=True, timeout=6, stream=True) as get_resp:
                return _validate_image_response(get_resp)
        try:
            return _validate_image_response(resp)
        finally:
            resp.close()
    except Exception:
        return False

def resolve_canonical_url(url: str) -> str:
    """Resolves Google Search Grounding redirect URLs to their final destination URL."""
    if not url or not isinstance(url, str):
        return ""
    cleaned = url.strip()
    host = urllib.parse.urlsplit(cleaned).netloc.lower()
    if "vertexaisearch.cloud.google.com" in host or "grounding-api-redirect" in cleaned.lower():
        try:
            with requests.get(
                cleaned,
                headers={"User-Agent": BROWSER_UA},
                allow_redirects=True,
                timeout=8,
                stream=True,
            ) as resp:
                if resp.url and resp.url.startswith(("http://", "https://")):
                    return resp.url
        except Exception:
            pass
    return cleaned

def _is_recipe_type(node_type) -> bool:
    if isinstance(node_type, str):
        return node_type.split("/")[-1].lower() == "recipe"
    if isinstance(node_type, list):
        return any(_is_recipe_type(item) for item in node_type)
    return False

def _extract_image_urls_from_value(val) -> List[str]:
    urls: List[str] = []
    if isinstance(val, str):
        if val.strip():
            urls.append(val.strip())
    elif isinstance(val, dict):
        for key in ("url", "contentUrl"):
            u = val.get(key)
            if isinstance(u, str) and u.strip():
                urls.append(u.strip())
                break
    elif isinstance(val, list):
        for item in val:
            urls.extend(_extract_image_urls_from_value(item))
    return urls

def _extract_recipe_jsonld_images(node) -> List[str]:
    urls: List[str] = []
    if isinstance(node, list):
        for item in node:
            urls.extend(_extract_recipe_jsonld_images(item))
    elif isinstance(node, dict):
        if _is_recipe_type(node.get("@type")) and "image" in node:
            urls.extend(_extract_image_urls_from_value(node.get("image")))
        for key, value in node.items():
            if key != "image" and isinstance(value, (dict, list)):
                urls.extend(_extract_recipe_jsonld_images(value))
    return urls

def _extract_image_candidates_from_html(html: str, base_url: str) -> List[str]:
    """Extracts prioritized candidate image URLs from Recipe JSON-LD, Open Graph, and Twitter Card tags."""
    soup = BeautifulSoup(html, "html.parser")
    raw_candidates: List[str] = []

    for script in soup.find_all("script", attrs={"type": re.compile(r"application/ld\+json", re.I)}):
        script_text = script.string or script.get_text() or ""
        if not script_text.strip():
            continue
        try:
            payload = json.loads(script_text)
            raw_candidates.extend(_extract_recipe_jsonld_images(payload))
        except Exception:
            continue

    meta_specs = [
        ("property", "og:image"),
        ("property", "og:image:secure_url"),
        ("name", "og:image"),
        ("name", "twitter:image"),
        ("property", "twitter:image"),
    ]
    for attr_name, attr_val in meta_specs:
        for tag in soup.find_all("meta", attrs={attr_name: re.compile(f"^{re.escape(attr_val)}$", re.I)}):
            content = tag.get("content")
            if isinstance(content, str) and content.strip():
                raw_candidates.append(content.strip())

    resolved: List[str] = []
    seen = set()
    for cand in raw_candidates:
        abs_url = urllib.parse.urljoin(base_url, cand)
        if abs_url not in seen:
            seen.add(abs_url)
            resolved.append(abs_url)
    return resolved

def extract_image_from_webpage(url: str) -> Optional[str]:
    """Scrapes the recipe source webpage for a valid Recipe JSON-LD, og:image, or twitter:image URL."""
    if not url or not re.match(r"^https?://[^/]+/.+", url.strip()):
        return None
    try:
        headers = {**_headers_for_url(url), "Accept": "text/html,application/xhtml+xml"}
        resp = requests.get(url.strip(), headers=headers, allow_redirects=True, timeout=8)
        content_type = (resp.headers.get("Content-Type") or "").lower()
        if resp.status_code != 200 or "html" not in content_type:
            return None
        for candidate in _extract_image_candidates_from_html(resp.text, resp.url):
            if is_valid_image_url(candidate):
                return candidate
    except Exception:
        return None
    return None

def _dish_search_queries(title_en: str) -> List[str]:
    """Builds ordered search query variants from an English dish title."""
    cleaned = (title_en or "").strip()
    if not cleaned:
        return []

    queries: List[str] = [cleaned]
    for inner in re.findall(r"\(([^)]+)\)", cleaned):
        inner_clean = inner.strip()
        if inner_clean and inner_clean not in queries:
            queries.append(inner_clean)

    stripped = re.sub(r"\([^)]*\)", "", cleaned).strip()
    changed = True
    while changed and stripped:
        changed = False
        lower = stripped.lower()
        for prefix in LEADING_CUISINE_PREFIXES:
            if lower.startswith(prefix + " "):
                stripped = stripped[len(prefix):].strip(" ,-")
                changed = True
                break

    if stripped and stripped not in queries:
        queries.append(stripped)
    return queries

def search_themealdb_image(title_en: str) -> Optional[str]:
    """Searches TheMealDB public API for a matching dish photograph."""
    for q in _dish_search_queries(title_en):
        try:
            resp = requests.get(
                "https://www.themealdb.com/api/json/v1/1/search.php",
                params={"s": q},
                headers={"User-Agent": BROWSER_UA},
                timeout=6,
            )
            if resp.status_code != 200:
                continue
            meals = resp.json().get("meals") or []
            q_words = {
                w for w in slugify(q).split("-") if len(w) > 2 and w not in STOPWORDS
            }
            ranked_meals = []
            for meal in meals:
                meal_name = meal.get("strMeal") or ""
                meal_words = {
                    w for w in slugify(meal_name).split("-") if len(w) > 2 and w not in STOPWORDS
                }
                overlap = len(q_words & meal_words)
                if q_words and overlap == 0:
                    continue
                exact = 1 if slugify(meal_name) == slugify(q) else 0
                ranked_meals.append((exact, overlap, meal.get("strMealThumb")))

            ranked_meals.sort(key=lambda item: (item[0], item[1]), reverse=True)
            for _, _, thumb in ranked_meals:
                if is_valid_image_url(thumb):
                    return thumb
        except Exception:
            continue
    return None

def search_wikipedia_image(title_en: str) -> Optional[str]:
    """Searches English Wikipedia PageImages API for a verified dish photograph."""
    base_queries = _dish_search_queries(title_en)
    if not base_queries:
        return None
    search_queries = list(base_queries)
    core_query = base_queries[-1]
    dish_hint = f"{core_query} dish"
    if dish_hint not in search_queries:
        search_queries.append(dish_hint)

    for q in search_queries:
        try:
            resp = requests.get(
                "https://en.wikipedia.org/w/api.php",
                params={
                    "action": "query",
                    "format": "json",
                    "generator": "search",
                    "gsrsearch": q,
                    "gsrlimit": 5,
                    "prop": "pageimages",
                    "pithumbsize": 1000,
                },
                headers={"User-Agent": WIKI_UA},
                timeout=8,
            )
            if resp.status_code != 200:
                continue
            pages = list((resp.json().get("query") or {}).get("pages", {}).values())
            pages.sort(key=lambda p: p.get("index", 99))

            preferred = [
                p for p in pages
                if not (p.get("title") or "").lower().endswith(("cuisine", "people"))
            ]
            fallback = [p for p in pages if p not in preferred]
            for page in preferred + fallback:
                src = (page.get("thumbnail") or {}).get("source")
                if not src:
                    continue
                cleaned_src = _clean_tracking_params(src)
                if is_valid_image_url(cleaned_src):
                    return cleaned_src
        except Exception:
            continue
    return None

def search_wikimedia_commons_image(title_en: str) -> Optional[str]:
    """Searches Wikimedia Commons files for a high-resolution dish photograph."""
    queries = _dish_search_queries(title_en)
    if not queries:
        return None
    core_query = queries[-1]
    try:
        resp = requests.get(
            "https://commons.wikimedia.org/w/api.php",
            params={
                "action": "query",
                "format": "json",
                "generator": "search",
                "gsrnamespace": 6,
                "gsrsearch": f"{core_query} food",
                "gsrlimit": 5,
                "prop": "imageinfo",
                "iiprop": "url|mime|size",
                "iiurlwidth": 1000,
            },
            headers={"User-Agent": WIKI_UA},
            timeout=8,
        )
        if resp.status_code != 200:
            return None
        pages = list((resp.json().get("query") or {}).get("pages", {}).values())
        pages.sort(key=lambda p: p.get("index", 99))
        for page in pages:
            info = (page.get("imageinfo") or [{}])[0]
            if info.get("mime") not in {"image/jpeg", "image/png", "image/webp"}:
                continue
            if (info.get("width") or 0) < 400:
                continue
            cand = _clean_tracking_params(info.get("thumburl") or info.get("url") or "")
            if is_valid_image_url(cand):
                return cand
    except Exception:
        return None
    return None

def resolve_dish_thumbnail(recipe: RecipeExtraction) -> Optional[str]:
    """Resolves a verified, working food photograph across AI output, source HTML, and public APIs."""
    if is_valid_image_url(recipe.thumbnail):
        resolved = recipe.thumbnail.strip()
        print("      [Image] Validated provided thumbnail URL.")
        return resolved

    layers = (
        ("source webpage", lambda: extract_image_from_webpage(recipe.source_url)),
        ("TheMealDB", lambda: search_themealdb_image(recipe.title_en)),
        ("Wikipedia PageImages", lambda: search_wikipedia_image(recipe.title_en)),
        ("Wikimedia Commons", lambda: search_wikimedia_commons_image(recipe.title_en)),
    )
    for layer_name, resolver in layers:
        candidate = resolver()
        if candidate:
            print(f"      [Image] Resolved via {layer_name}: {candidate}")
            return candidate

    print(f"      [WARN] Could not resolve a verified thumbnail image for '{recipe.title_en}'.")
    return None

def slugify(text: str) -> str:
    """Converts text (including accented international and Vietnamese characters) into a clean URL slug."""
    text = unicodedata.normalize('NFKD', text)
    text = text.replace('đ', 'd').replace('Đ', 'D')
    text = text.encode('ascii', 'ignore').decode('utf-8')
    text = re.sub(r'[^\w\s-]', '', text).strip().lower()
    return re.sub(r'[-\s]+', '-', text)

def get_existing_dish_titles(limit: int = 80) -> List[str]:
    """Fetches recent existing dish titles from MongoDB to avoid redundant AI suggestions."""
    titles: List[str] = []
    try:
        cursor = dishes_col.find(
            {"deleted": False},
            {"title": 1, "slug": 1}
        ).sort("createdAt", -1).limit(limit)
        for doc in cursor:
            title_list = doc.get("title") or []
            en_title = next((t.get("data") for t in title_list if t.get("lang") == "en" and t.get("data")), None)
            vi_title = next((t.get("data") for t in title_list if t.get("lang") == "vi" and t.get("data")), None)
            chosen = en_title or vi_title or doc.get("slug")
            if chosen:
                titles.append(chosen)
    except Exception as e:
        print(f"  [WARN] Could not fetch existing dish titles: {e}")
    return titles

def _build_exclusion_clause(exclude_titles: Optional[List[str]]) -> str:
    """Formats the list of existing dish titles to exclude from AI generation."""
    if not exclude_titles:
        return ""
    return f"Do NOT suggest any dishes already in our database, such as: {', '.join(exclude_titles[-80:])}"

def _extract_display_title(doc: dict) -> str:
    """Returns a readable display title preferring English, then Vietnamese, then slug."""
    title_list = doc.get("title") or []
    return (
        next((t.get("data") for t in title_list if t.get("lang") == "en" and t.get("data")), None)
        or next((t.get("data") for t in title_list if t.get("lang") == "vi" and t.get("data")), None)
        or doc.get("slug", "")
    )

# ==============================================================================
# Tier 1 & Tier 2: AI Recipe Discovery
# ==============================================================================
def discover_recipes_with_search_grounding(
    category_keyword: str,
    count: int,
    exclude_titles: Optional[List[str]] = None
) -> List[RecipeExtraction]:
    """Tier 1: Try Gemini with Google Search Grounding."""
    if not ai_client:
        return []

    print(f"  [Tier 1: Search Grounding] Searching Google for '{category_keyword}'...")
    exclude_clause = _build_exclusion_clause(exclude_titles)
    search_prompt = f"""
    Search Google for authentic, popular recipes from around the world matching: '{category_keyword}'.
    Look up real recipe pages from reputable global culinary websites (such as seriouseats.com, bbcgoodfood.com, allrecipes.com, recipetineats.com, justonecookbook.com, bonappetit.com, foodnetwork.com, epicurious.com, cookpad.com, and authentic regional culinary sites).
    Select {count} distinct, iconic or everyday beloved dishes for this cuisine/category.
    {exclude_clause}
    For each dish, extract full bilingual data (both English and Vietnamese):
    - Title (EN & VI) and short description (EN & VI)
    - Authentic recipe source_url and high-resolution thumbnail image URL (leave thumbnail null if not directly visible in search results; the pipeline automatically extracts/resolves verified dish photos)
    - Step-by-step cooking instructions in semantic HTML (<p>, <h4>, <ul>, <li>) in both EN & VI
    - Realistic preparation_time and cooking_time (in minutes), difficulty_level ('EASY', 'MEDIUM', or 'HARD')
    - Applicable meal_categories strictly from: ['BREAKFAST', 'LUNCH', 'DINNER', 'SNACK', 'FRIED', 'GRILLED', 'STEAMED', 'BAKED', 'SOUP', 'NOODLE', 'RICE', 'SEAFOOD', 'MEAT', 'VEGETARIAN', 'TRADITIONAL', 'STREET_FOOD']
    - Bilingual tags (EN & VI) including the cuisine/country of origin
    - Complete bilingual ingredients with quantities, units, notes, and category strictly from: ['PROTEIN FOODS', 'VEGETABLES', 'GRAINS', 'FATS AND OILS', 'CONDIMENTS AND SAUCES', 'HERBS AND SPICES', 'MILK AND DAIRY', 'FRUIT', 'NUTS AND SEEDS', 'OTHER']
    """

    search_response = ai_client.models.generate_content(
        model=GEMINI_MODEL,
        contents=search_prompt,
        config=types.GenerateContentConfig(
            tools=[types.Tool(google_search=types.GoogleSearch())],
            temperature=0.3,
        ),
    )

    grounding_uris = []
    if search_response.candidates and search_response.candidates[0].grounding_metadata:
        chunks = search_response.candidates[0].grounding_metadata.grounding_chunks or []
        for c in chunks:
            if hasattr(c, "web") and c.web and c.web.uri:
                grounding_uris.append(c.web.uri)

    grounding_hint = f"\nRelevant source URLs discovered: {', '.join(grounding_uris[:6])}" if grounding_uris else ""

    format_prompt = f"""
    Convert the following grounded recipe search results into the required JSON schema.
    Ensure `source_url` accurately contains the original web URL where the recipe is found.
    {grounding_hint}

    Content:
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

def discover_recipes_direct_ai(
    category_keyword: str,
    count: int,
    exclude_titles: Optional[List[str]] = None
) -> List[RecipeExtraction]:
    """Tier 2: Direct Gemini generation (consumes 0 Search Grounding quota, single call)."""
    if not ai_client:
        return []

    print(f"  [Tier 2: Direct AI] Generating '{category_keyword}' recipe from culinary knowledge base...")
    exclude_clause = _build_exclusion_clause(exclude_titles)
    prompt = f"""
    You are an expert international chef and culinary data architect with deep knowledge of world cuisines.
    Generate {count} authentic, beloved recipes matching the cuisine or category: '{category_keyword}'.
    {exclude_clause}
    Requirements:
    1. Authentic English title and natural Vietnamese translation title.
    2. Cite a specific reputable international recipe page URL (not a bare homepage, e.g. https://www.seriouseats.com/..., https://www.bbcgoodfood.com/recipes/..., https://www.recipetineats.com/..., https://www.justonecookbook.com/...).
    3. Short descriptions in both EN and VI mentioning the dish's flavor and culinary origin.
    4. Detailed step-by-step cooking instructions in semantic HTML (<p>, <h4>, <ul>, <li>) in both EN and VI.
    5. Set thumbnail to null unless you are certain of a real, working direct image URL (the ingestion pipeline automatically resolves verified dish photographs).
    6. Realistic preparation_time and cooking_time in minutes.
    7. Difficulty: 'EASY', 'MEDIUM', or 'HARD'.
    8. Meal categories strictly from: ['BREAKFAST', 'LUNCH', 'DINNER', 'SNACK', 'FRIED', 'GRILLED', 'STEAMED', 'BAKED', 'SOUP', 'NOODLE', 'RICE', 'SEAFOOD', 'MEAT', 'VEGETARIAN', 'TRADITIONAL', 'STREET_FOOD'], plus bilingual tags (EN & VI) including the cuisine/country of origin.
    9. Complete bilingual ingredient list with realistic quantities, units (g, kg, ml, tbsp, tsp, cup, piece, clove), notes, and nutritional categories strictly from: ['PROTEIN FOODS', 'VEGETABLES', 'GRAINS', 'FATS AND OILS', 'CONDIMENTS AND SAUCES', 'HERBS AND SPICES', 'MILK AND DAIRY', 'FRUIT', 'NUTS AND SEEDS', 'OTHER'].
    """

    structured_response = ai_client.models.generate_content(
        model=GEMINI_MODEL,
        contents=prompt,
        config=types.GenerateContentConfig(
            response_mime_type="application/json",
            response_schema=RecipeBatch,
            temperature=0.7,
        ),
    )

    batch = RecipeBatch.model_validate_json(structured_response.text)
    return batch.recipes

# ==============================================================================
# Tier 3: Curated World Recipe Catalog Fallback
# ==============================================================================
def get_curated_pool() -> List[RecipeExtraction]:
    """Tier 3: Built-in authentic world recipes pool ensuring at least 1 dish is ALWAYS ingested."""
    return [
        RecipeExtraction(
            title_vi="Mì Ý Carbonara",
            title_en="Spaghetti Carbonara",
            short_desc_vi="Món mì Ý kinh điển vùng Rome với sốt trứng béo ngậy, phô mai Pecorino Romano mặn mà và thịt má heo muối guanciale giòn thơm.",
            short_desc_en="Classic Roman pasta tossed in a silky emulsion of egg yolks, Pecorino Romano cheese, cracked black pepper, and crisp cured pork.",
            content_vi="<h4>Bước 1: Sơ chế nguyên liệu</h4><p>Thái thịt ba rọi xông khói hoặc guanciale thành miếng vuông nhỏ. Đánh tan lòng đỏ trứng với phô mai Pecorino Romano bào mịn và nhiều tiêu đen xay.</p><h4>Bước 2: Luộc mì</h4><p>Luộc spaghetti trong nước sôi pha muối ít hơn hướng dẫn 1 phút để mì đạt độ dai al dente, giữ lại khoảng 150ml nước luộc mì.</p><h4>Bước 3: Áp chảo thịt</h4><p>Chiên thịt ở lửa vừa đến khi mỡ tiết ra và miếng thịt vàng giòn, tắt bếp cho chảo nguội bớt.</p><h4>Bước 4: Trộn sốt</h4><p>Cho mì nóng vào chảo thịt cùng một ít nước luộc mì, sau đó trút hỗn hợp trứng phô mai vào đảo nhanh tay ngoài bếp để tạo sốt sánh mịn mà không bị vón cục. Rắc thêm phô mai và tiêu đen.</p>",
            content_en="<h4>Step 1: Prep Ingredients</h4><p>Dice cured pork cheek or pancetta into small lardons. Whisk egg yolks with finely grated Pecorino Romano and generous freshly cracked black pepper.</p><h4>Step 2: Cook Pasta</h4><p>Boil spaghetti in salted water until just al dente, reserving 150 ml of starchy pasta cooking water before draining.</p><h4>Step 3: Crisp the Pork</h4><p>Render the pork over medium heat until golden and crisp, then remove the pan from the heat.</p><h4>Step 4: Emulsify Sauce</h4><p>Toss hot pasta with the pork and a splash of pasta water, then stir in the egg-cheese mixture off the heat, adding pasta water as needed to form a glossy sauce.</p>",
            thumbnail="https://thumb.wikimedia.org/wikipedia/commons/thumb/3/33/Espaguetis_carbonara.jpg/1280px-Espaguetis_carbonara.jpg",
            source_url="https://en.wikipedia.org/wiki/Carbonara",
            preparation_time=10,
            cooking_time=15,
            difficulty_level="MEDIUM",
            meal_categories=["LUNCH", "DINNER", "NOODLE", "MEAT", "TRADITIONAL"],
            tags=["italian", "món ý", "italy", "pasta", "carbonara", "spaghetti", "mì ý"],
            ingredients=[
                ParsedIngredient(name_vi="Mì Ý Spaghetti", name_en="Spaghetti", quantity=400.0, measure="g", note="dry pasta", category="GRAINS"),
                ParsedIngredient(name_vi="Thịt má heo muối hoặc ba rọi xông khói", name_en="Guanciale or pancetta", quantity=150.0, measure="g", note="diced", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Lòng đỏ trứng gà", name_en="Egg yolks", quantity=4.0, measure="piece", note="room temperature", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Phô mai Pecorino Romano", name_en="Pecorino Romano cheese", quantity=80.0, measure="g", note="finely grated", category="MILK AND DAIRY"),
                ParsedIngredient(name_vi="Tiêu đen xay", name_en="Black pepper", quantity=5.0, measure="g", note="freshly cracked", category="HERBS AND SPICES"),
            ]
        ),
        RecipeExtraction(
            title_vi="Cà ri gà chiên xù Nhật Bản",
            title_en="Chicken Katsu Curry",
            short_desc_vi="Ức gà chiên xù panko giòn rụm ăn kèm cơm Nhật dẻo thơm và nước sốt cà ri nâu sánh đậm vị ngọt dịu.",
            short_desc_en="Crispy panko-breaded chicken cutlet served over steamed rice with rich, mildly spiced Japanese brown curry sauce.",
            content_vi="<h4>Bước 1: Nấu sốt cà ri</h4><p>Xào thơm hành tây thái lát, cà rốt và khoai tây với dầu ăn. Đổ nước dùng vào ninh mềm khoảng 15 phút, sau đó thả viên cà ri Nhật vào khuấy tan đến khi sánh.</p><h4>Bước 2: Chuẩn bị gà</h4><p>Dần mỏng ức gà, nêm muối tiêu. Lần lượt áo gà qua bột mì, trứng đánh tan và vụn bánh mì panko.</p><h4>Bước 3: Chiên giòn gà</h4><p>Chiên gà ngập dầu ở 175°C khoảng 4-5 phút mỗi mặt đến khi vàng giòn rồi vớt ra để ráo, thái lát vừa ăn.</p><h4>Bước 4: Trình bày</h4><p>Xới cơm nóng ra đĩa, xếp gà katsu lên trên và chan đều sốt cà ri nóng hổi.</p>",
            content_en="<h4>Step 1: Simmer Curry Sauce</h4><p>Sauté sliced onions, carrots, and diced potatoes until softened. Add water or stock, simmer 15 minutes, then stir in Japanese curry roux blocks until glossy and thick.</p><h4>Step 2: Bread the Chicken</h4><p>Pound chicken breasts to even thickness and season with salt and pepper. Dredge in flour, dip in beaten egg, and coat thoroughly with panko breadcrumbs.</p><h4>Step 3: Deep Fry</h4><p>Fry at 175°C for 4-5 minutes until crisp and golden brown. Rest briefly, then slice into strips.</p><h4>Step 4: Serve</h4><p>Plate steamed short-grain rice, top with sliced chicken katsu, and ladle hot curry sauce alongside.</p>",
            thumbnail="https://thumb.wikimedia.org/wikipedia/commons/thumb/c/cc/%E3%81%A8%E3%82%93%E3%81%8B%E3%82%89%E4%BA%AD_%E6%AD%A6%E8%94%B5%E9%87%8E%E3%83%96%E3%83%A9%E3%83%83%E3%82%AF%E3%81%8B%E3%81%A4%E3%82%AB%E3%83%AC%E3%83%BC_-_1.jpg/1280px-%E3%81%A8%E3%82%93%E3%81%8B%E3%82%89%E4%BA%AD_%E6%AD%A6%E8%94%B5%E9%87%8E%E3%83%96%E3%83%A9%E3%83%83%E3%82%AF%E3%81%8B%E3%81%A4%E3%82%AB%E3%83%AC%E3%83%BC_-_1.jpg",
            source_url="https://en.wikipedia.org/wiki/Katsu_curry",
            preparation_time=20,
            cooking_time=25,
            difficulty_level="MEDIUM",
            meal_categories=["LUNCH", "DINNER", "FRIED", "RICE", "MEAT"],
            tags=["japanese", "món nhật", "japan", "katsu curry", "cà ri nhật", "panko chicken"],
            ingredients=[
                ParsedIngredient(name_vi="Ức gà", name_en="Chicken breast", quantity=400.0, measure="g", note="boneless, skinless", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Vụn bánh mì Panko", name_en="Panko breadcrumbs", quantity=100.0, measure="g", note=None, category="GRAINS"),
                ParsedIngredient(name_vi="Viên sốt cà ri Nhật", name_en="Japanese curry roux", quantity=100.0, measure="g", note="mild or medium", category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Hành tây", name_en="Onion", quantity=1.0, measure="piece", note="thinly sliced", category="VEGETABLES"),
                ParsedIngredient(name_vi="Cà rốt", name_en="Carrot", quantity=100.0, measure="g", note="sliced", category="VEGETABLES"),
                ParsedIngredient(name_vi="Cơm trắng hạt ngắn", name_en="Short-grain white rice", quantity=300.0, measure="g", note="steamed", category="GRAINS"),
                ParsedIngredient(name_vi="Dầu ăn", name_en="Cooking oil", quantity=250.0, measure="ml", note="for frying", category="FATS AND OILS"),
            ]
        ),
        RecipeExtraction(
            title_vi="Bánh Taco thịt heo nướng al Pastor",
            title_en="Tacos al Pastor",
            short_desc_vi="Món taco đường phố nổi tiếng của Mexico với thịt heo ướp ớt guajillo và dứa thơm lừng, kẹp bánh ngô mềm và hành ngò.",
            short_desc_en="Iconic Mexican street tacos filled with smoky chili-and-achiote marinated pork, sweet grilled pineapple, onion, and cilantro.",
            content_vi="<h4>Bước 1: Làm sốt ướp</h4><p>Ngâm mềm ớt khô guajillo rồi xay nhuyễn cùng bột điều achiote, tỏi, giấm táo, nước dứa và thì là Ai Cập.</p><h4>Bước 2: Ướp thịt heo</h4><p>Thái mỏng thịt nạc vai heo, trộn đều với sốt ướp và để thấm ít nhất 2 tiếng trong tủ lạnh.</p><h4>Bước 3: Áp chảo hoặc nướng</h4><p>Nướng hoặc áp chảo lửa lớn thịt heo cùng dứa thái lát cho xém cạnh thơm lừng, sau đó băm nhỏ vừa ăn.</p><h4>Bước 4: Cuốn taco</h4><p>Làm nóng bánh tortilla ngô, xếp thịt nướng, dứa nướng, hành tây băm, ngò rí và vắt thêm chanh tươi.</p>",
            content_en="<h4>Step 1: Blend Marinade</h4><p>Rehydrate dried guajillo chilies and blend with achiote paste, garlic, cider vinegar, pineapple juice, oregano, and cumin until smooth.</p><h4>Step 2: Marinate Pork</h4><p>Thinly slice pork shoulder, coat evenly in the marinade, and refrigerate for at least 2 hours.</p><h4>Step 3: Sear and Chop</h4><p>Grill or sear the pork and pineapple slices over high heat until charred at the edges, then chop into bite-sized pieces.</p><h4>Step 4: Assemble Tacos</h4><p>Warm corn tortillas on a skillet, fill with pork and charred pineapple, and top with diced white onion, cilantro, and fresh lime.</p>",
            thumbnail="https://upload.wikimedia.org/wikipedia/commons/4/4a/Tacos-al-Pastor.jpg",
            source_url="https://en.wikipedia.org/wiki/Al_pastor",
            preparation_time=25,
            cooking_time=20,
            difficulty_level="MEDIUM",
            meal_categories=["LUNCH", "DINNER", "GRILLED", "MEAT", "STREET_FOOD", "TRADITIONAL"],
            tags=["mexican", "món mexico", "mexico", "tacos", "al pastor", "street food", "bánh taco"],
            ingredients=[
                ParsedIngredient(name_vi="Thịt nạc vai heo", name_en="Pork shoulder", quantity=600.0, measure="g", note="thinly sliced", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Bánh Tortilla bắp", name_en="Corn tortillas", quantity=8.0, measure="piece", note="warmed", category="GRAINS"),
                ParsedIngredient(name_vi="Dứa tươi", name_en="Fresh pineapple", quantity=200.0, measure="g", note="sliced", category="FRUIT"),
                ParsedIngredient(name_vi="Ớt khô Guajillo", name_en="Dried guajillo chilies", quantity=3.0, measure="piece", note="seeded and soaked", category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Sốt điều Achiote", name_en="Achiote paste", quantity=30.0, measure="g", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Hành tây trắng", name_en="White onion", quantity=1.0, measure="piece", note="finely diced", category="VEGETABLES"),
                ParsedIngredient(name_vi="Ngò rí", name_en="Fresh cilantro", quantity=30.0, measure="g", note="chopped", category="HERBS AND SPICES"),
            ]
        ),
        RecipeExtraction(
            title_vi="Gà nấu bơ cà ri Ấn Độ",
            title_en="Butter Chicken (Murgh Makhani)",
            short_desc_vi="Thịt gà ướp sữa chua nướng thơm quyện trong nước sốt cà chua bơ béo mịn và gia vị garam masala nồng ấm.",
            short_desc_en="Tender spiced chicken simmered in a velvety, aromatic tomato-butter cream sauce finished with fenugreek leaves.",
            content_vi="<h4>Bước 1: Ướp và áp chảo gà</h4><p>Ướp thịt đùi gà cắt khối với sữa chua không đường, gừng tỏi băm, nước cốt chanh và garam masala trong 30 phút rồi áp chảo xém cạnh.</p><h4>Bước 2: Nấu nền sốt cà chua</h4><p>Đun chảy bơ, xào thơm hành tây, gừng tỏi, thêm cà chua nghiền, bột ớt Kashmir và đun nhỏ lửa 15 phút rồi xay nhuyễn.</p><h4>Bước 3: Hoàn thiện sốt bơ</h4><p>Đổ sốt trở lại nồi, thêm gà đã áp chảo, kem tươi whipping cream, một ít bơ và lá cỏ cà ri khô kasuri methi.</p><h4>Bước 4: Thưởng thức</h4><p>Đun liu riu thêm 8 phút cho gà thấm vị, dùng nóng với bánh mì naan hoặc cơm basmati.</p>",
            content_en="<h4>Step 1: Marinate and Sear Chicken</h4><p>Marinate boneless chicken thighs in yogurt, ginger-garlic paste, lemon juice, and garam masala for 30 minutes, then sear until lightly charred.</p><h4>Step 2: Simmer Tomato Base</h4><p>Melt butter in a pan, sauté onions and aromatics, add crushed tomatoes and Kashmiri chili powder, simmer 15 minutes, and blend smooth.</p><h4>Step 3: Finish with Cream and Butter</h4><p>Return sauce to the pan with seared chicken, heavy cream, extra butter, and crushed dried fenugreek leaves.</p><h4>Step 4: Simmer and Serve</h4><p>Simmer gently for 8 minutes until chicken is cooked through. Serve hot with naan or basmati rice.</p>",
            thumbnail="https://thumb.wikimedia.org/wikipedia/commons/thumb/4/41/Butter_Chicken_%26_Butter_Naan_-_Home_-_Chandigarh_-_India_-_0006.jpg/1280px-Butter_Chicken_%26_Butter_Naan_-_Home_-_Chandigarh_-_India_-_0006.jpg",
            source_url="https://en.wikipedia.org/wiki/Butter_chicken",
            preparation_time=30,
            cooking_time=30,
            difficulty_level="MEDIUM",
            meal_categories=["LUNCH", "DINNER", "MEAT", "TRADITIONAL"],
            tags=["indian", "món ấn độ", "india", "butter chicken", "curry", "cà ri gà bơ"],
            ingredients=[
                ParsedIngredient(name_vi="Thịt đùi gà không xương", name_en="Boneless chicken thighs", quantity=600.0, measure="g", note="cut into bite-sized pieces", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Sữa chua không đường", name_en="Plain yogurt", quantity=120.0, measure="g", note="for marinade", category="MILK AND DAIRY"),
                ParsedIngredient(name_vi="Cà chua nghiền", name_en="Crushed tomatoes", quantity=400.0, measure="g", note=None, category="VEGETABLES"),
                ParsedIngredient(name_vi="Bơ lạt", name_en="Unsalted butter", quantity=60.0, measure="g", note=None, category="MILK AND DAIRY"),
                ParsedIngredient(name_vi="Kem tươi whipping cream", name_en="Heavy cream", quantity=100.0, measure="ml", note=None, category="MILK AND DAIRY"),
                ParsedIngredient(name_vi="Bột gia vị Garam Masala", name_en="Garam masala", quantity=10.0, measure="g", note=None, category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Gừng và tỏi băm", name_en="Ginger garlic paste", quantity=20.0, measure="g", note=None, category="HERBS AND SPICES"),
            ]
        ),
        RecipeExtraction(
            title_vi="Phở xào Pad Thai",
            title_en="Pad Thai",
            short_desc_vi="Hủ tiếu gạo Thái Lan xào lửa lớn với tôm tươi, trứng, giá đỗ và nước sốt me chua ngọt mặn hài hòa rắc đậu phộng rang.",
            short_desc_en="Classic Thai stir-fried rice noodles glazed in tangy tamarind-fish sauce with shrimp, egg, garlic chives, and crushed peanuts.",
            content_vi="<h4>Bước 1: Ngâm hủ tiếu và pha sốt</h4><p>Ngâm sợi bánh phở khô trong nước ấm cho vừa mềm. Khuấy đều cốt me chua, nước mắm và đường thốt nốt thành sốt Pad Thai.</p><h4>Bước 2: Xào tôm và đậu phụ</h4><p>Phi thơm hành tím và tỏi băm, xào nhanh tôm tươi và đậu phụ chiên cắt hạt lựu cho săn lại.</p><h4>Bước 3: Xào mì và trứng</h4><p>Đẩy tôm sang một bên, đập trứng vào đảo tơi rồi cho sợi phở và sốt me vào đảo đều ở lửa lớn đến khi sợi mì ngấm sốt.</p><h4>Bước 4: Hoàn thành</h4><p>Cho giá đỗ và hẹ cắt khúc vào đảo nhanh 30 giây, trút ra đĩa rắc đậu phộng rang giã dập và kèm chanh tươi.</p>",
            content_en="<h4>Step 1: Soak Noodles and Mix Sauce</h4><p>Soak dried flat rice noodles in warm water until pliable. Whisk tamarind paste, fish sauce, and palm sugar into a balanced sweet-tangy sauce.</p><h4>Step 2: Sear Shrimp and Tofu</h4><p>Stir-fry minced shallots and garlic in hot oil, then sear shrimp and diced firm tofu until lightly golden.</p><h4>Step 3: Stir-fry Noodles and Egg</h4><p>Push ingredients aside, scramble eggs in the wok, then toss in drained noodles and tamarind sauce over high heat until absorbed.</p><h4>Step 4: Garnish and Serve</h4><p>Fold in bean sprouts and garlic chives for 30 seconds. Serve topped with crushed roasted peanuts and fresh lime wedges.</p>",
            thumbnail="https://thumb.wikimedia.org/wikipedia/commons/thumb/3/39/Phat_Thai_kung_Chang_Khien_street_stall.jpg/1280px-Phat_Thai_kung_Chang_Khien_street_stall.jpg",
            source_url="https://en.wikipedia.org/wiki/Pad_thai",
            preparation_time=20,
            cooking_time=15,
            difficulty_level="EASY",
            meal_categories=["LUNCH", "DINNER", "NOODLE", "SEAFOOD", "STREET_FOOD", "TRADITIONAL"],
            tags=["thai", "món thái", "thailand", "pad thai", "rice noodles", "street food"],
            ingredients=[
                ParsedIngredient(name_vi="Bánh phở khô bản dẹt", name_en="Flat rice noodles", quantity=250.0, measure="g", note="soaked until pliable", category="GRAINS"),
                ParsedIngredient(name_vi="Tôm tươi bóc vỏ", name_en="Peeled shrimp", quantity=250.0, measure="g", note="deveined", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Trứng gà", name_en="Eggs", quantity=2.0, measure="piece", note=None, category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Cốt me chua", name_en="Tamarind paste", quantity=45.0, measure="g", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Nước mắm", name_en="Fish sauce", quantity=35.0, measure="ml", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Giá đỗ", name_en="Bean sprouts", quantity=100.0, measure="g", note=None, category="VEGETABLES"),
                ParsedIngredient(name_vi="Đậu phộng rang", name_en="Roasted peanuts", quantity=40.0, measure="g", note="crushed", category="NUTS AND SEEDS"),
            ]
        ),
        RecipeExtraction(
            title_vi="Cơm trộn Bibimbap Hàn Quốc",
            title_en="Korean Bibimbap",
            short_desc_vi="Tô cơm trộn rực rỡ sắc màu với thịt bò xào bulgogi, các loại rau củ xào namul, trứng ốp la và tương ớt gochujang cay ngọt.",
            short_desc_en="Vibrant Korean mixed rice bowl topped with seasoned sautéed vegetables, savory beef, a sunny-side-up egg, and spicy gochujang sauce.",
            content_vi="<h4>Bước 1: Xào các loại rau namul</h4><p>Chần cải bó xôi rồi trộn dầu mè, tỏi. Xào riêng từng loại cà rốt thái sợi, bí ngòi và nấm đông cô với chút muối và dầu mè.</p><h4>Bước 2: Xào thịt bò</h4><p>Ướp thịt bò thái mỏng với nước tương, đường, tỏi băm, dầu mè rồi xào nhanh trên lửa lớn cho chín mềm.</p><h4>Bước 3: Chiên trứng và pha sốt</h4><p>Chiên trứng ốp la lòng đào. Trộn tương ớt gochujang với mật ong, giấm gạo, dầu mè và mè rang.</p><h4>Bước 4: Trình bày</h4><p>Cho cơm nóng vào tô, xếp đều thịt bò và các loại rau củ xung quanh, đặt trứng ốp la ở giữa và chan sốt gochujang khi ăn.</p>",
            content_en="<h4>Step 1: Prepare Vegetable Toppings</h4><p>Blanch spinach and dress with sesame oil and garlic. Separately sauté julienned carrots, zucchini, and shiitake mushrooms with salt and sesame oil.</p><h4>Step 2: Cook Savory Beef</h4><p>Marinate thinly sliced beef in soy sauce, sugar, garlic, and sesame oil, then stir-fry over high heat until browned.</p><h4>Step 3: Fry Egg and Mix Sauce</h4><p>Fry an egg sunny-side up. Whisk gochujang paste with honey, rice vinegar, sesame oil, and toasted sesame seeds.</p><h4>Step 4: Assemble Bowl</h4><p>Place warm steamed rice in a bowl, arrange beef and vegetables radially on top, crown with the egg, and serve with gochujang sauce.</p>",
            thumbnail="https://thumb.wikimedia.org/wikipedia/commons/thumb/4/44/Dolsot-bibimbap.jpg/1280px-Dolsot-bibimbap.jpg",
            source_url="https://en.wikipedia.org/wiki/Bibimbap",
            preparation_time=25,
            cooking_time=20,
            difficulty_level="MEDIUM",
            meal_categories=["LUNCH", "DINNER", "RICE", "MEAT", "TRADITIONAL"],
            tags=["korean", "món hàn", "korea", "bibimbap", "cơm trộn hàn quốc", "gochujang"],
            ingredients=[
                ParsedIngredient(name_vi="Cơm trắng dẻo", name_en="Steamed short-grain rice", quantity=350.0, measure="g", note="warm", category="GRAINS"),
                ParsedIngredient(name_vi="Thịt bò thăn thái mỏng", name_en="Thinly sliced beef sirloin", quantity=200.0, measure="g", note=None, category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Cải bó xôi", name_en="Spinach", quantity=120.0, measure="g", note=None, category="VEGETABLES"),
                ParsedIngredient(name_vi="Cà rốt và bí ngòi", name_en="Carrot and zucchini", quantity=200.0, measure="g", note="julienned", category="VEGETABLES"),
                ParsedIngredient(name_vi="Tương ớt Gochujang", name_en="Gochujang chili paste", quantity=45.0, measure="g", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Dầu mè", name_en="Sesame oil", quantity=20.0, measure="ml", note=None, category="FATS AND OILS"),
                ParsedIngredient(name_vi="Trứng gà", name_en="Egg", quantity=2.0, measure="piece", note="sunny-side up", category="PROTEIN FOODS"),
            ]
        ),
        RecipeExtraction(
            title_vi="Bò hầm rượu vang kiểu Pháp",
            title_en="Beef Bourguignon",
            short_desc_vi="Thịt bò hầm chậm trong rượu vang đỏ Burgundy cùng thịt xông khói, cà rốt, nấm mỡ và thảo mộc thơm nồng đậm đà.",
            short_desc_en="Classic French slow-braised beef stew simmered in rich red Burgundy wine with bacon lardons, carrots, pearl onions, and mushrooms.",
            content_vi="<h4>Bước 1: Áp chảo thịt xông khói và bò</h4><p>Áp chảo thịt ba rọi xông khói cho ra mỡ rồi vớt ra. Thấm khô thịt bò cắt khối, áp chảo vàng đều các mặt trong phần mỡ đó.</p><h4>Bước 2: Xào rau củ thơm</h4><p>Cho cà rốt, hành tây, tỏi và sốt cà chua cô đặc vào nồi đảo đều, rắc thêm chút bột mì để tạo độ sánh.</p><h4>Bước 3: Hầm rượu vang</h4><p>Đổ rượu vang đỏ và nước dùng bò ngập thịt, thêm lá nguyệt quế và cỏ xạ hương. Đậy nắp hầm nhỏ lửa trong 2 đến 2,5 tiếng cho thịt mềm tan.</p><h4>Bước 4: Thêm nấm và hoàn tất</h4><p>Xào vàng nấm mỡ với bơ rồi cho vào nồi hầm trong 15 phút cuối. Nêm lại muối tiêu và rắc ngò tây trước khi dùng.</p>",
            content_en="<h4>Step 1: Brown Bacon and Beef</h4><p>Crisp bacon lardons in a Dutch oven and set aside. Pat beef chuck cubes dry and sear in batches until deeply browned.</p><h4>Step 2: Sauté Aromatics</h4><p>Cook carrots, onions, garlic, and tomato paste in the pot, then dust lightly with flour to thicken the base.</p><h4>Step 3: Braise in Red Wine</h4><p>Pour in red Burgundy wine and beef stock to cover the meat, add thyme and bay leaf, cover, and braise gently for 2 to 2.5 hours until fork-tender.</p><h4>Step 4: Fold in Mushrooms</h4><p>Sauté quartered mushrooms in butter and fold into the stew for the last 15 minutes. Garnish with chopped parsley.</p>",
            thumbnail="https://thumb.wikimedia.org/wikipedia/commons/thumb/2/29/Beef_bourguignon_NYT.jpg/1280px-Beef_bourguignon_NYT.jpg",
            source_url="https://en.wikipedia.org/wiki/Beef_bourguignon",
            preparation_time=30,
            cooking_time=150,
            difficulty_level="HARD",
            meal_categories=["DINNER", "MEAT", "TRADITIONAL"],
            tags=["french", "món pháp", "france", "beef bourguignon", "bò hầm rượu vang", "stew"],
            ingredients=[
                ParsedIngredient(name_vi="Thịt bắp hoặc nạm bò", name_en="Beef chuck", quantity=800.0, measure="g", note="cut into 4cm cubes", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Rượu vang đỏ", name_en="Dry red wine", quantity=500.0, measure="ml", note="Burgundy or Pinot Noir", category="OTHER"),
                ParsedIngredient(name_vi="Thịt ba rọi xông khói", name_en="Bacon lardons", quantity=150.0, measure="g", note="diced", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Nấm mỡ", name_en="Cremini mushrooms", quantity=250.0, measure="g", note="quartered", category="VEGETABLES"),
                ParsedIngredient(name_vi="Cà rốt", name_en="Carrots", quantity=200.0, measure="g", note="chunked", category="VEGETABLES"),
                ParsedIngredient(name_vi="Bơ lạt", name_en="Butter", quantity=30.0, measure="g", note=None, category="MILK AND DAIRY"),
                ParsedIngredient(name_vi="Cỏ xạ hương và lá nguyệt quế", name_en="Thyme and bay leaves", quantity=5.0, measure="g", note=None, category="HERBS AND SPICES"),
            ]
        ),
        RecipeExtraction(
            title_vi="Moussaka cà tím nướng kiểu Hy Lạp",
            title_en="Greek Moussaka",
            short_desc_vi="Món đút lò truyền thống của Hy Lạp xếp lớp cà tím nướng, thịt cừu hoặc bò hầm sốt cà chua quế thơm và lớp sốt kem béchamel béo vàng.",
            short_desc_en="Traditional Greek layered casserole of roasted eggplant, spiced meat and tomato ragu, and a rich golden béchamel topping.",
            content_vi="<h4>Bước 1: Nướng cà tím</h4><p>Cắt lát cà tím dày 1cm, rắc muối để ráo bớt nước đắng, phết dầu ô liu rồi nướng vàng hai mặt.</p><h4>Bước 2: Nấu sốt thịt</h4><p>Xào hành tỏi và thịt băm với cà chua nghiền, chút quế, oregano và rượu vang trắng đến khi sốt sánh đậm đà.</p><h4>Bước 3: Làm sốt kem Béchamel</h4><p>Đun chảy bơ với bột mì, từ từ rót sữa ấm khuấy đều đến khi sệt mịn, tắt bếp rồi khuấy thêm lòng đỏ trứng, nhục đậu khấu và phô mai bào.</p><h4>Bước 4: Xếp lớp và nướng</h4><p>Xếp lớp cà tím, sốt thịt, cà tím và phủ sốt béchamel lên trên cùng. Nướng ở 180°C khoảng 40 phút đến khi mặt bánh vàng ruộm.</p>",
            content_en="<h4>Step 1: Roast Eggplant</h4><p>Slice eggplant into 1 cm rounds, salt lightly to draw out moisture, pat dry, brush with olive oil, and roast until tender and golden.</p><h4>Step 2: Simmer Meat Ragu</h4><p>Brown ground lamb or beef with onions and garlic, add crushed tomatoes, cinnamon, and oregano, and simmer until thick.</p><h4>Step 3: Whisk Béchamel</h4><p>Cook butter and flour into a roux, slowly whisk in warm milk until thick, then fold in egg yolks, grated cheese, and nutmeg off the heat.</p><h4>Step 4: Layer and Bake</h4><p>Layer roasted eggplant and meat sauce in a baking dish, pour béchamel over the top, and bake at 180°C for 40 minutes until golden and bubbling.</p>",
            thumbnail="https://thumb.wikimedia.org/wikipedia/commons/thumb/a/a8/MussakasMeMelitsanesKePatates01.JPG/1280px-MussakasMeMelitsanesKePatates01.JPG",
            source_url="https://en.wikipedia.org/wiki/Moussaka",
            preparation_time=40,
            cooking_time=50,
            difficulty_level="HARD",
            meal_categories=["LUNCH", "DINNER", "BAKED", "MEAT", "TRADITIONAL"],
            tags=["greek", "mediterranean", "món hy lạp", "greece", "moussaka", "eggplant"],
            ingredients=[
                ParsedIngredient(name_vi="Cà tím", name_en="Eggplant", quantity=700.0, measure="g", note="sliced into rounds", category="VEGETABLES"),
                ParsedIngredient(name_vi="Thịt cừu hoặc thịt bò băm", name_en="Ground lamb or beef", quantity=500.0, measure="g", note=None, category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Cà chua nghiền", name_en="Crushed tomatoes", quantity=400.0, measure="g", note=None, category="VEGETABLES"),
                ParsedIngredient(name_vi="Sữa tươi nguyên kem", name_en="Whole milk", quantity=500.0, measure="ml", note="for béchamel", category="MILK AND DAIRY"),
                ParsedIngredient(name_vi="Bơ lạt", name_en="Unsalted butter", quantity=50.0, measure="g", note=None, category="MILK AND DAIRY"),
                ParsedIngredient(name_vi="Dầu ô liu", name_en="Olive oil", quantity=45.0, measure="ml", note=None, category="FATS AND OILS"),
                ParsedIngredient(name_vi="Bột quế và oregano", name_en="Ground cinnamon and oregano", quantity=8.0, measure="g", note=None, category="HERBS AND SPICES"),
            ]
        ),
        RecipeExtraction(
            title_vi="Trứng chần sốt cà chua Shakshuka",
            title_en="Shakshuka",
            short_desc_vi="Món ăn sáng phổ biến khắp Trung Đông và Bắc Phi với trứng chần chín tới trong chảo sốt cà chua, ớt chuông, thì là và ớt paprika cay ấm.",
            short_desc_en="Beloved Middle Eastern and North African skillet dish of eggs poached directly in a spiced, fragrant tomato and bell pepper sauce.",
            content_vi="<h4>Bước 1: Xào rau củ và gia vị</h4><p>Đun nóng dầu ô liu, xào mềm hành tây và ớt chuông đỏ thái hạt lựu. Thêm tỏi, bột thì là Ai Cập, paprika và bột ớt đảo thơm.</p><h4>Bước 2: Ninh sốt cà chua</h4><p>Đổ cà chua băm hoặc cà chua đóng hộp vào chảo, nêm muối tiêu và đun nhỏ lửa khoảng 10-12 phút cho sốt sánh lại.</p><h4>Bước 3: Chần trứng trong sốt</h4><p>Tạo các hốc nhỏ trên mặt sốt rồi đập từng quả trứng vào. Đậy nắp đun lửa nhỏ 5-7 phút đến khi lòng trắng chín nhưng lòng đào vẫn dẻo mềm.</p><h4>Bước 4: Hoàn thành</h4><p>Rắc ngò rí, ngò tây và chút phô mai feta nếu thích, ăn nóng kèm bánh mì nướng giòn.</p>",
            content_en="<h4>Step 1: Sauté Peppers and Spices</h4><p>Heat olive oil in a skillet and sauté diced onion and red bell pepper until soft. Stir in minced garlic, cumin, smoked paprika, and chili flakes.</p><h4>Step 2: Simmer Tomato Sauce</h4><p>Pour in crushed or diced tomatoes, season with salt and pepper, and simmer for 10-12 minutes until thickened.</p><h4>Step 3: Poach Eggs</h4><p>Make small wells in the sauce and crack an egg into each. Cover and simmer gently for 5-7 minutes until egg whites set and yolks remain runny.</p><h4>Step 4: Garnish</h4><p>Top with fresh cilantro or parsley and crumbled feta if desired. Serve straight from the skillet with crusty bread.</p>",
            thumbnail="https://thumb.wikimedia.org/wikipedia/commons/thumb/1/18/Shakshuka_by_Calliopejen1.jpg/1280px-Shakshuka_by_Calliopejen1.jpg",
            source_url="https://en.wikipedia.org/wiki/Shakshouka",
            preparation_time=10,
            cooking_time=20,
            difficulty_level="EASY",
            meal_categories=["BREAKFAST", "LUNCH", "DINNER", "VEGETARIAN", "TRADITIONAL"],
            tags=["middle eastern", "north african", "shakshuka", "món trung đông", "vegetarian", "eggs"],
            ingredients=[
                ParsedIngredient(name_vi="Trứng gà", name_en="Eggs", quantity=4.0, measure="piece", note=None, category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Cà chua băm nhuyễn", name_en="Crushed tomatoes", quantity=400.0, measure="g", note=None, category="VEGETABLES"),
                ParsedIngredient(name_vi="Ớt chuông đỏ", name_en="Red bell pepper", quantity=1.0, measure="piece", note="diced", category="VEGETABLES"),
                ParsedIngredient(name_vi="Hành tây", name_en="Onion", quantity=1.0, measure="piece", note="diced", category="VEGETABLES"),
                ParsedIngredient(name_vi="Dầu ô liu", name_en="Olive oil", quantity=30.0, measure="ml", note=None, category="FATS AND OILS"),
                ParsedIngredient(name_vi="Bột thì là và Paprika", name_en="Ground cumin and paprika", quantity=10.0, measure="g", note=None, category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Ngò rí", name_en="Fresh cilantro", quantity=20.0, measure="g", note="chopped", category="HERBS AND SPICES"),
            ]
        ),
        RecipeExtraction(
            title_vi="Cơm hải sản Paella Tây Ban Nha",
            title_en="Seafood Paella",
            short_desc_vi="Món cơm chảo vàng óng nghệ tây saffron đặc sản Valencia, thấm đẫm nước dùng hải sản cùng tôm, vẹm xanh và mực tươi.",
            short_desc_en="Iconic Spanish saffron rice cooked in a wide shallow pan with rich seafood stock, plump shrimp, mussels, and tender squid.",
            content_vi="<h4>Bước 1: Chuẩn bị nền sofrito</h4><p>Xào mềm hành tây, tỏi, ớt chuông và cà chua băm với dầu ô liu và bột ớt hun khói Tây Ban Nha đến khi sánh lại.</p><h4>Bước 2: Xào gạo và ủ nghệ tây</h4><p>Cho gạo hạt tròn Bomba vào đảo đều cho thấm dầu. Hòa nhụy hoa nghệ tây saffron vào nước dùng hải sản nóng rồi chan đều vào chảo.</p><h4>Bước 3: Nấu cơm không khuấy</h4><p>Dàn đều gạo và nấu lửa vừa khoảng 15 phút mà không khuấy để hình thành lớp cơm cháy giòn socarrat ở đáy chảo.</p><h4>Bước 4: Thêm hải sản</h4><p>Xếp tôm, vẹm xanh và mực lên mặt cơm, nấu thêm 6-8 phút đến khi hải sản chín và vẹm mở vỏ. Trang trí chanh vàng và ngò tây.</p>",
            content_en="<h4>Step 1: Build the Sofrito</h4><p>Sauté onion, garlic, bell pepper, and grated tomato in olive oil with smoked paprika until thick and aromatic.</p><h4>Step 2: Toast Rice and Add Saffron Stock</h4><p>Stir in Spanish short-grain rice to coat with the sofrito. Infuse saffron threads into hot seafood stock and pour evenly over the rice.</p><h4>Step 3: Simmer Without Stirring</h4><p>Cook undisturbed over medium heat for about 15 minutes so the rice absorbs the broth and forms a crisp socarrat crust underneath.</p><h4>Step 4: Cook Seafood</h4><p>Nestle shrimp, mussels, and squid rings on top of the rice, cover loosely or cook 6-8 minutes until mussels open and shrimp turn pink. Serve with lemon wedges.</p>",
            thumbnail="https://thumb.wikimedia.org/wikipedia/commons/thumb/e/ed/01_Paella_Valenciana_original.jpg/1280px-01_Paella_Valenciana_original.jpg",
            source_url="https://en.wikipedia.org/wiki/Paella",
            preparation_time=20,
            cooking_time=30,
            difficulty_level="MEDIUM",
            meal_categories=["LUNCH", "DINNER", "RICE", "SEAFOOD", "TRADITIONAL"],
            tags=["spanish", "món tây ban nha", "spain", "paella", "seafood", "cơm hải sản"],
            ingredients=[
                ParsedIngredient(name_vi="Gạo hạt ngắn Tây Ban Nha", name_en="Bomba or short-grain rice", quantity=300.0, measure="g", note=None, category="GRAINS"),
                ParsedIngredient(name_vi="Tôm sú", name_en="Large shrimp", quantity=250.0, measure="g", note="shell-on or peeled", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Vẹm xanh", name_en="Mussels", quantity=300.0, measure="g", note="scrubbed and debearded", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Mực ống cắt khoanh", name_en="Squid rings", quantity=200.0, measure="g", note=None, category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Nhụy hoa nghệ tây Saffron", name_en="Saffron threads", quantity=1.0, measure="g", note="steeped in warm stock", category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Dầu ô liu", name_en="Olive oil", quantity=45.0, measure="ml", note=None, category="FATS AND OILS"),
                ParsedIngredient(name_vi="Chanh vàng", name_en="Lemon", quantity=1.0, measure="piece", note="cut into wedges", category="FRUIT"),
            ]
        ),
        RecipeExtraction(
            title_vi="Gà hầm Tagine kiểu Ma-rốc",
            title_en="Moroccan Chicken Tagine",
            short_desc_vi="Thịt gà hầm mềm trong nồi đất cùng chanh muối, ô liu xanh, nghệ tây, gừng và mật ong mang hương vị Bắc Phi quyến rũ.",
            short_desc_en="Fragrant North African braised chicken stewed with preserved lemon, green olives, warm spices, and sweet caramelized onions.",
            content_vi="<h4>Bước 1: Ướp và áp chảo gà</h4><p>Trộn đùi gà với gừng, nghệ, bột thì là, quế, tỏi băm và dầu ô liu rồi áp chảo vàng đều hai mặt.</p><h4>Bước 2: Xào hành và gia vị</h4><p>Cho hành tây thái lát vào nồi đảo mềm với nghệ tây và nước dùng gà.</p><h4>Bước 3: Hầm trong nồi Tagine</h4><p>Đặt gà trở lại nồi, đậy nắp hầm nhỏ lửa khoảng 35-40 phút đến khi thịt gà mềm nhừ và nước sốt sánh lại.</p><h4>Bước 4: Thêm ô liu và chanh muối</h4><p>Cho ô liu xanh và vỏ chanh muối thái lát vào đun thêm 5 phút. Rắc ngò rí và hạnh nhân rang lên trên, ăn kèm couscous.</p>",
            content_en="<h4>Step 1: Brown Spiced Chicken</h4><p>Coat chicken thighs with ground ginger, turmeric, cumin, cinnamon, garlic, and olive oil, then sear until golden brown.</p><h4>Step 2: Soften Onions</h4><p>Sauté sliced onions in the same pot until translucent and fragrant, then pour in chicken broth infused with saffron.</p><h4>Step 3: Slow Simmer</h4><p>Return chicken to the tagine or Dutch oven, cover, and simmer gently for 35-40 minutes until tender and flavorful.</p><h4>Step 4: Finish with Olives and Preserved Lemon</h4><p>Stir in pitted green olives and sliced preserved lemon peel for the last 5 minutes. Garnish with cilantro and toasted almonds, and serve with couscous.</p>",
            thumbnail="https://upload.wikimedia.org/wikipedia/commons/3/3a/Tajine-marocain-un-plat-varie-et-sain_%28cropped%29.jpg",
            source_url="https://en.wikipedia.org/wiki/Tajine",
            preparation_time=20,
            cooking_time=45,
            difficulty_level="MEDIUM",
            meal_categories=["LUNCH", "DINNER", "MEAT", "TRADITIONAL"],
            tags=["moroccan", "north african", "món ma-rốc", "morocco", "tagine", "chicken"],
            ingredients=[
                ParsedIngredient(name_vi="Đùi gà có xương", name_en="Bone-in chicken thighs", quantity=700.0, measure="g", note=None, category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Chanh muối Ma-rốc", name_en="Preserved lemon", quantity=1.0, measure="piece", note="rind thinly sliced", category="FRUIT"),
                ParsedIngredient(name_vi="Quả ô liu xanh", name_en="Green olives", quantity=100.0, measure="g", note="pitted", category="FRUIT"),
                ParsedIngredient(name_vi="Hành tây", name_en="Yellow onion", quantity=2.0, measure="piece", note="sliced", category="VEGETABLES"),
                ParsedIngredient(name_vi="Hạnh nhân lát rang", name_en="Toasted sliced almonds", quantity=30.0, measure="g", note="for garnish", category="NUTS AND SEEDS"),
                ParsedIngredient(name_vi="Nghệ, gừng và quế", name_en="Turmeric, ginger, and cinnamon", quantity=12.0, measure="g", note=None, category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Dầu ô liu", name_en="Olive oil", quantity=30.0, measure="ml", note=None, category="FATS AND OILS"),
            ]
        ),
        RecipeExtraction(
            title_vi="Đậu phụ Tứ Xuyên Mapo",
            title_en="Sichuan Mapo Tofu",
            short_desc_vi="Đậu phụ non mềm mượt hòa quyện cùng thịt băm, tương đậu cay Tứ Xuyên Doubanjiang và hạt tiêu hoa Tứ Xuyên tê nồng.",
            short_desc_en="Silky soft tofu braised in a fiery, umami-packed Sichuan chili bean sauce with minced pork and numbing Sichuan peppercorns.",
            content_vi="<h4>Bước 1: Sơ chế đậu phụ</h4><p>Cắt đậu phụ non thành khối vuông 2cm, chần nhẹ qua nước sôi pha chút muối rồi vớt ra để ráo giúp đậu không bị nát.</p><h4>Bước 2: Xào thịt và tương đậu cay</h4><p>Xào tơi thịt heo băm với gừng, tỏi băm và gốc hành lá. Thêm tương đậu cay Tứ Xuyên Doubanjiang và đậu đen lên men đảo đều đến khi lên màu đỏ đẹp.</p><h4>Bước 3: Om đậu phụ</h4><p>Đổ nước dùng gà và chút nước tương vào đun sôi, nhẹ nhàng thả đậu phụ vào om 5 phút để thấm gia vị.</p><h4>Bước 4: Tạo độ sánh</h4><p>Hòa tan bột bắp với nước rồi rưới từ từ vào chảo để sốt sánh bóng. Rắc bột hoa tiêu Tứ Xuyên và hành lá lên trên, ăn nóng với cơm trắng.</p>",
            content_en="<h4>Step 1: Blanch Tofu</h4><p>Cut soft tofu into 2 cm cubes and gently simmer in lightly salted water for 2 minutes, then drain to keep cubes intact.</p><h4>Step 2: Brown Pork and Chili Bean Paste</h4><p>Stir-fry minced pork until crisp, then add ginger, garlic, scallions, Sichuan doubanjiang paste, and fermented black beans until fragrant and red.</p><h4>Step 3: Braise Tofu</h4><p>Add stock and soy sauce, bring to a simmer, and gently slide in the tofu cubes. Simmer for 5 minutes.</p><h4>Step 4: Thicken and Finish</h4><p>Stir in a cornstarch slurry until the sauce turns glossy. Sprinkle freshly ground Sichuan peppercorns and green onions on top, and serve with steamed rice.</p>",
            thumbnail="https://thumb.wikimedia.org/wikipedia/commons/thumb/8/88/Chen_Mapo_Tofu.jpg/1280px-Chen_Mapo_Tofu.jpg",
            source_url="https://en.wikipedia.org/wiki/Mapo_tofu",
            preparation_time=15,
            cooking_time=15,
            difficulty_level="EASY",
            meal_categories=["LUNCH", "DINNER", "MEAT", "TRADITIONAL"],
            tags=["chinese", "sichuan", "món trung hoa", "china", "mapo tofu", "đậu phụ tứ xuyên"],
            ingredients=[
                ParsedIngredient(name_vi="Đậu phụ non", name_en="Soft tofu", quantity=400.0, measure="g", note="cubed", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Thịt heo băm", name_en="Ground pork", quantity=150.0, measure="g", note=None, category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Tương đậu cay Tứ Xuyên", name_en="Doubanjiang chili bean paste", quantity=35.0, measure="g", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Hoa tiêu Tứ Xuyên", name_en="Sichuan peppercorns", quantity=5.0, measure="g", note="toasted and ground", category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Nước tương", name_en="Soy sauce", quantity=15.0, measure="ml", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Bột bắp", name_en="Cornstarch", quantity=10.0, measure="g", note="mixed with water", category="GRAINS"),
                ParsedIngredient(name_vi="Hành lá và tỏi", name_en="Scallions and garlic", quantity=40.0, measure="g", note="chopped", category="VEGETABLES"),
            ]
        ),
        RecipeExtraction(
            title_vi="Thịt hầm đậu đen Feijoada Brazil",
            title_en="Brazilian Feijoada",
            short_desc_vi="Món hầm quốc hồn quốc túy của Brazil kết hợp đậu đen ninh nhừ cùng thịt heo muối, xúc xích hun khói và hương lá nguyệt quế đậm đà.",
            short_desc_en="Brazil's hearty national black bean stew slow-simmered with smoked sausage, pork shoulder, garlic, and bay leaves.",
            content_vi="<h4>Bước 1: Ngâm đậu đen</h4><p>Rửa sạch đậu đen và ngâm nước qua đêm hoặc ít nhất 6 tiếng cho đậu nở đều.</p><h4>Bước 2: Áp chảo thịt và xúc xích</h4><p>Áp chảo vàng thịt vai heo cắt khối và xúc xích hun khói chorizo cắt lát, sau đó xào thơm hành tây và tỏi băm.</p><h4>Bước 3: Ninh đậu với thịt</h4><p>Cho đậu đen đã ngâm, các loại thịt, lá nguyệt quế và nước vào nồi lớn, đun sôi rồi hạ nhỏ lửa ninh khoảng 1,5 - 2 tiếng đến khi đậu mềm và nước sánh.</p><h4>Bước 4: Phục vụ</h4><p>Nêm muối tiêu vừa ăn, múc ra tô dùng nóng cùng cơm trắng, cải xanh xào tỏi và vài lát cam tươi.</p>",
            content_en="<h4>Step 1: Soak Black Beans</h4><p>Rinse dried black beans and soak overnight in plenty of cold water, then drain.</p><h4>Step 2: Brown Meats and Aromatics</h4><p>Sear cubed pork shoulder and sliced smoked sausage in a large pot until browned, then sauté chopped onion and minced garlic.</p><h4>Step 3: Slow Simmer</h4><p>Add soaked black beans, bay leaves, and water to cover. Simmer gently for 1.5 to 2 hours until the beans are tender and the broth is rich and creamy.</p><h4>Step 4: Serve</h4><p>Season to taste and serve hot over steamed rice alongside sautéed collard greens and fresh orange slices.</p>",
            thumbnail="https://upload.wikimedia.org/wikipedia/commons/8/84/Feijoada_01.jpg",
            source_url="https://en.wikipedia.org/wiki/Feijoada_(Brazilian_dish)",
            preparation_time=25,
            cooking_time=120,
            difficulty_level="MEDIUM",
            meal_categories=["LUNCH", "DINNER", "MEAT", "TRADITIONAL"],
            tags=["brazilian", "latin american", "món brazil", "brazil", "feijoada", "black bean stew"],
            ingredients=[
                ParsedIngredient(name_vi="Đậu đen khô", name_en="Dried black beans", quantity=450.0, measure="g", note="soaked overnight", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Thịt nạc vai heo", name_en="Pork shoulder", quantity=400.0, measure="g", note="cubed", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Xúc xích hun khói", name_en="Smoked sausage", quantity=250.0, measure="g", note="sliced", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Hành tây", name_en="Onion", quantity=1.0, measure="piece", note="chopped", category="VEGETABLES"),
                ParsedIngredient(name_vi="Tỏi", name_en="Garlic", quantity=6.0, measure="clove", note="minced", category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Lá nguyệt quế", name_en="Bay leaves", quantity=3.0, measure="piece", note=None, category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Cam tươi", name_en="Fresh orange", quantity=1.0, measure="piece", note="sliced for serving", category="FRUIT"),
            ]
        ),
        RecipeExtraction(
            title_vi="Phở bò truyền thống",
            title_en="Traditional Vietnamese Beef Pho",
            short_desc_vi="Nước dùng xương bò ninh nhiều giờ thơm nồng quế, hồi, thảo quả kết hợp bánh phở mềm dai, thịt bò tái chín và rau thơm.",
            short_desc_en="Aromatic Vietnamese noodle soup featuring slow-simmered spiced beef bone broth, silky rice noodles, tender beef slices, and fresh herbs.",
            content_vi="<h4>Bước 1: Ninh xương bò</h4><p>Chần sơ xương bò qua nước sôi rồi rửa sạch. Ninh xương nhỏ lửa khoảng 4 tiếng, thường xuyên vớt bọt để nước dùng trong ngọt.</p><h4>Bước 2: Nướng hành gừng và rang gia vị</h4><p>Nướng thơm hành tây và gừng già rồi cạo vỏ cho vào nồi nước dùng. Rang hoa hồi, quế thanh, thảo quả cho dậy mùi, bọc túi vải thả vào nồi.</p><h4>Bước 3: Trụng bánh phở và thịt bò</h4><p>Nêm nước dùng với nước mắm ngon, muối và chút đường phèn. Trụng bánh phở cho vào tô, xếp thịt nạm bò và thịt thăn bò thái mỏng lên trên.</p><h4>Bước 4: Hoàn thiện</h4><p>Chan nước dùng đang sôi sùng sục ngập mặt thịt, rắc hành lá, ngò rí và thưởng thức nóng cùng giá đỗ, rau quế và chanh ớt.</p>",
            content_en="<h4>Step 1: Simmer Bone Broth</h4><p>Parboil beef bones for 5 minutes and rinse clean. Simmer gently in fresh water for 4 hours, skimming foam to keep the broth clear.</p><h4>Step 2: Char Aromatics and Toast Spices</h4><p>Char onion and ginger until fragrant, then add to the pot alongside toasted star anise, cinnamon stick, and black cardamom wrapped in spice cloth.</p><h4>Step 3: Assemble Bowls</h4><p>Season broth with fish sauce, salt, and rock sugar. Blanch flat rice noodles, place into bowls, and layer cooked brisket and thinly sliced raw beef sirloin on top.</p><h4>Step 4: Ladle Boiling Broth</h4><p>Pour piping hot broth over the beef to cook it gently. Garnish with scallions and cilantro, and serve with bean sprouts, Thai basil, and lime.</p>",
            thumbnail="https://thumb.wikimedia.org/wikipedia/commons/thumb/5/52/Bowl_of_Meatball_pho.jpg/1280px-Bowl_of_Meatball_pho.jpg",
            source_url="https://en.wikipedia.org/wiki/Pho",
            preparation_time=30,
            cooking_time=240,
            difficulty_level="HARD",
            meal_categories=["BREAKFAST", "LUNCH", "DINNER", "SOUP", "NOODLE", "MEAT", "TRADITIONAL"],
            tags=["vietnamese", "món việt", "vietnam", "pho", "phở bò", "beef noodle soup"],
            ingredients=[
                ParsedIngredient(name_vi="Xương ống bò", name_en="Beef marrow bones", quantity=1000.0, measure="g", note="parboiled", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Thịt thăn bò", name_en="Beef sirloin", quantity=300.0, measure="g", note="paper-thin slices", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Bánh phở tươi", name_en="Flat rice noodles", quantity=500.0, measure="g", note=None, category="GRAINS"),
                ParsedIngredient(name_vi="Gừng tươi và hành tây", name_en="Fresh ginger and onion", quantity=200.0, measure="g", note="charred", category="VEGETABLES"),
                ParsedIngredient(name_vi="Quế, hoa hồi và thảo quả", name_en="Cinnamon, star anise, and cardamom", quantity=20.0, measure="g", note="toasted", category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Nước mắm", name_en="Fish sauce", quantity=45.0, measure="ml", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Hành lá và rau quế", name_en="Scallions and Thai basil", quantity=50.0, measure="g", note="for garnish", category="HERBS AND SPICES"),
            ]
        ),
    ]

# ==============================================================================
# Database Ingestion Logic
# ==============================================================================
def get_or_create_ingredient(ing: ParsedIngredient, author_id: str = "system-crawler") -> Tuple[str, str]:
    """Finds existing ingredient by slug or title, or auto-creates it in Ingredients collection."""
    en_slug = slugify(ing.name_en)
    vi_slug = slugify(ing.name_vi)
    slug = en_slug or vi_slug
    candidate_slugs = [s for s in (en_slug, vi_slug) if s]

    now = datetime.now(timezone.utc)

    if candidate_slugs:
        existing = ingredients_col.find_one({"slug": {"$in": candidate_slugs}, "deleted": False})
        if existing:
            return str(existing["_id"]), existing.get("slug") or slug

    title_patterns = [
        {"title.data": {"$regex": f"^{re.escape(name.strip())}$", "$options": "i"}}
        for name in (ing.name_en, ing.name_vi)
        if name and name.strip()
    ]
    existing_by_title = (
        ingredients_col.find_one({"$or": title_patterns, "deleted": False})
        if title_patterns else None
    )
    if existing_by_title:
        return str(existing_by_title["_id"]), existing_by_title.get("slug") or slug

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
    print(f"    [+] Auto-created Ingredient: {ing.name_en} / {ing.name_vi} (slug: {slug})")
    return str(new_doc["_id"]), slug

def find_and_link_related_dishes(
    dish_id: ObjectId,
    dish_slug: str,
    title_en: str,
    title_vi: str,
    meal_categories: List[str],
    tags: List[str],
    limit: int = 4
) -> List[str]:
    """
    Searches the MongoDB database for 2-4 related dishes (referenced by their ObjectIds).
    1. Searches dishes sharing tags in MongoDB.
    2. Searches dishes sharing culinary keywords from title/slug in MongoDB.
    3. Searches dishes sharing mealCategories in MongoDB.
    4. Falls back to recent dishes if fewer than 2.
    5. Bidirectionally links by adding dish_id to each matched dish's relatedDishes array.
    """
    target_id_str = str(dish_id)
    matched_ids: List[str] = []
    seen_ids = {target_id_str}

    print(f"    [~] Searching database for dishes related to '{dish_slug}'...")

    # Strategy 1: Match existing dishes sharing tags
    if tags:
        clean_tags = [t.strip().lower() for t in tags if t and len(t.strip()) > 1]
        for tag in clean_tags:
            if len(matched_ids) >= limit:
                break
            try:
                for doc in dishes_col.find({"deleted": False, "tags": tag}).limit(limit):
                    did = str(doc["_id"])
                    if did not in seen_ids and doc.get("slug") != dish_slug:
                        matched_ids.append(did)
                        seen_ids.add(did)
                        print(f"      [+] Found in DB by tag '{tag}' -> {_extract_display_title(doc)} (ID: {did})")
                        if len(matched_ids) >= limit:
                            break
            except Exception:
                pass

    # Strategy 2: Match existing dishes by culinary keywords from English and Vietnamese titles
    if len(matched_ids) < limit:
        raw_words = slugify(title_en).split('-') + slugify(title_vi).split('-')
        title_words = list(dict.fromkeys(
            w for w in raw_words if len(w) > 2 and w not in STOPWORDS
        ))
        for kw in title_words:
            if len(matched_ids) >= limit:
                break
            try:
                for doc in dishes_col.find({"deleted": False, "slug": re.compile(re.escape(kw), re.IGNORECASE)}).limit(limit):
                    did = str(doc["_id"])
                    if did not in seen_ids and doc.get("slug") != dish_slug:
                        matched_ids.append(did)
                        seen_ids.add(did)
                        print(f"      [+] Found in DB by keyword '{kw}' -> {_extract_display_title(doc)} (ID: {did})")
                        if len(matched_ids) >= limit:
                            break
            except Exception:
                pass

    # Strategy 3: Match existing dishes by mealCategories
    if len(matched_ids) < limit and meal_categories:
        needed = limit - len(matched_ids)
        for cat in meal_categories:
            if len(matched_ids) >= limit:
                break
            try:
                for doc in dishes_col.find({"deleted": False, "mealCategories": cat}).limit(needed * 2):
                    did = str(doc["_id"])
                    if did not in seen_ids and doc.get("slug") != dish_slug:
                        matched_ids.append(did)
                        seen_ids.add(did)
                        print(f"      [+] Found in DB by category '{cat}' -> {_extract_display_title(doc)} (ID: {did})")
                        if len(matched_ids) >= limit:
                            break
            except Exception:
                pass

    # Strategy 4: Fallback to recent dishes
    if len(matched_ids) < 2:
        needed = limit - len(matched_ids)
        try:
            for doc in dishes_col.find({"deleted": False}).sort("createdAt", -1).limit(needed * 3):
                did = str(doc["_id"])
                if did not in seen_ids and doc.get("slug") != dish_slug:
                    matched_ids.append(did)
                    seen_ids.add(did)
                    print(f"      [+] Found in DB by recent fallback -> {_extract_display_title(doc)} (ID: {did})")
                    if len(matched_ids) >= limit:
                        break
        except Exception:
            pass

    # Bidirectional link: update each related dish document in the database
    for rel_id_str in matched_ids:
        try:
            if ObjectId.is_valid(rel_id_str):
                dishes_col.update_one(
                    {"_id": ObjectId(rel_id_str)},
                    {"$addToSet": {"relatedDishes": target_id_str}}
                )
        except Exception as e:
            print(f"      [WARN] Could not update bidirectional link for {rel_id_str}: {e}")

    return matched_ids

def ingest_recipe(recipe: RecipeExtraction, author_id: str = "github-actions-crawler") -> bool:
    """Inserts the recipe document into MongoDB with sourceUrl, ingredients, and relatedDishes."""
    en_slug = slugify(recipe.title_en)
    vi_slug = slugify(recipe.title_vi)
    dish_slug = en_slug or vi_slug
    if not dish_slug:
        print("  [WARN] Recipe produced an empty slug. Skipping.")
        return False

    candidate_slugs = list(dict.fromkeys(s for s in (en_slug, vi_slug) if s))
    canonical_source_url = resolve_canonical_url(recipe.source_url)
    recipe.source_url = canonical_source_url
    stripped_url = canonical_source_url.strip()
    or_conditions = [{"slug": {"$in": candidate_slugs}}]
    # Only deduplicate on sourceUrl when it points to a specific path beyond the root domain
    if re.match(r"^https?://[^/]+/.+", stripped_url):
        or_conditions.append({"sourceUrl": stripped_url})

    existing = dishes_col.find_one({"$or": or_conditions, "deleted": False})
    if existing:
        print(f"  [-] Dish already in database ('{dish_slug}'). Skipping duplicate.")
        return False

    print(f"\n  [>] Ingesting Dish: {recipe.title_en} / {recipe.title_vi}")
    print(f"      Source URL: {recipe.source_url}")
    resolved_thumbnail = resolve_dish_thumbnail(recipe)
    print(f"      Thumbnail: {resolved_thumbnail}")

    new_dish_id = ObjectId()

    mapped_ingredients = []
    for ing in recipe.ingredients:
        ing_id, ing_slug = get_or_create_ingredient(ing, author_id=author_id)
        mapped_ingredients.append({
            "ingredientId": ing_id,
            "slug": ing_slug,
            "quantity": ing.quantity,
            "note": ing.note
        })

    # Search database for related dishes and establish bidirectional links
    related_dish_ids = find_and_link_related_dishes(
        dish_id=new_dish_id,
        dish_slug=dish_slug,
        title_en=recipe.title_en,
        title_vi=recipe.title_vi,
        meal_categories=recipe.meal_categories,
        tags=recipe.tags,
        limit=4
    )

    now = datetime.now(timezone.utc)

    dish_doc = {
        "_id": new_dish_id,
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
        "thumbnail": resolved_thumbnail,
        "videos": [],
        "ingredients": mapped_ingredients,
        "relatedDishes": related_dish_ids,
        "labels": ["AUTO_CRAWLED", "RESILIENT_PIPELINE"],
        "sourceUrl": recipe.source_url,
        "deleted": False,
        "createdAt": now,
        "updatedAt": now,
        "createdBy": author_id,
        "updatedBy": author_id
    }

    dishes_col.insert_one(dish_doc)
    print(f"  [SUCCESS] Ingested '{recipe.title_en}' / '{recipe.title_vi}' (ID: {new_dish_id}) with {len(mapped_ingredients)} ingredients and {len(related_dish_ids)} related dishes linked.\n")
    return True

# ==============================================================================
# Main Resilient Pipeline
# ==============================================================================
def main():
    print("=" * 60)
    print("Starting Resilient World Recipe Ingestion Job (what-to-eat)")
    print(f"Target limit: {DAILY_CRAWL_LIMIT} dishes")
    print(f"Configured cuisines/categories ({len(CRAWL_CATEGORIES)}): {CRAWL_CATEGORIES}")
    print("=" * 60)

    total_inserted = 0
    categories_to_crawl = list(CRAWL_CATEGORIES)
    random.shuffle(categories_to_crawl)
    exclude_titles = get_existing_dish_titles(limit=80)

    # --------------------------------------------------------------------------
    # Step 1 & 2: Try AI Search Grounding, then Direct AI Generation
    # --------------------------------------------------------------------------
    for idx, category in enumerate(categories_to_crawl):
        if total_inserted >= DAILY_CRAWL_LIMIT:
            break

        remaining = DAILY_CRAWL_LIMIT - total_inserted
        remaining_cats = len(categories_to_crawl) - idx
        fetch_count = min(remaining, max(1, math.ceil(remaining / remaining_cats)), 5)
        recipes = []

        # Attempt Tier 1: Search Grounding
        try:
            recipes = discover_recipes_with_search_grounding(
                category,
                count=fetch_count,
                exclude_titles=exclude_titles,
            )
        except Exception as e_grounding:
            print(f"  [Notice] Tier 1 search grounding bypassed ({e_grounding})")

        # If Tier 1 failed or returned 0, fallback to Tier 2: Direct AI
        if not recipes and ai_client:
            try:
                recipes = discover_recipes_direct_ai(
                    category,
                    count=fetch_count,
                    exclude_titles=exclude_titles,
                )
            except Exception as e_direct:
                print(f"  [Notice] Tier 2 direct AI bypassed ({e_direct})")

        # Ingest whatever recipes were gathered
        for recipe in recipes:
            if total_inserted >= DAILY_CRAWL_LIMIT:
                break
            if ingest_recipe(recipe):
                total_inserted += 1
                new_title = recipe.title_en or recipe.title_vi
                if new_title:
                    exclude_titles.append(new_title)
            time.sleep(1)

    # --------------------------------------------------------------------------
    # Step 3: Guaranteed Fallback from Curated Pool
    # --------------------------------------------------------------------------
    if total_inserted < DAILY_CRAWL_LIMIT:
        print(f"\n[Tier 3 Fallback] Target not reached ({total_inserted}/{DAILY_CRAWL_LIMIT}). Drawing from Curated World Pool...")
        curated_recipes = get_curated_pool()
        random.shuffle(curated_recipes)
        for recipe in curated_recipes:
            if total_inserted >= DAILY_CRAWL_LIMIT:
                break
            if ingest_recipe(recipe):
                total_inserted += 1

    print("=" * 60)
    print(f"Job completed successfully. Total new dishes ingested: {total_inserted}")
    print("=" * 60)

    if total_inserted == 0:
        print("[WARN] All available recipes already exist in database.", file=sys.stderr)

if __name__ == "__main__":
    main()
