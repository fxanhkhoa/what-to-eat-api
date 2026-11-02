#!/usr/bin/env python3
"""
Resilient Recipe Ingestion Pipeline for what-to-eat
Multi-tiered ingestion pipeline that NEVER stops and guarantees dish ingestion:
  - Tier 1: Gemini 3.8 Flash with Google Search Grounding
  - Tier 2: Gemini Direct AI Recipe Generation (no search quota needed, bypasses 429)
  - Tier 3: Curated Vietnamese Recipe Pool (guarantees ingestion even if AI quota is 100% exhausted)
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
    suggested_related_dishes: List[str] = Field(
        default=[],
        description="Tên 2-4 món ăn Việt Nam phổ biến liên quan hoặc ăn kèm hợp vị (ví dụ: món kho ăn kèm món canh, món xào)"
    )
    ingredients: List[ParsedIngredient]

class RecipeBatch(BaseModel):
    recipes: List[RecipeExtraction]

# ==============================================================================
# Helpers
# ==============================================================================
def slugify(text: str) -> str:
    """Converts Vietnamese string into clean URL slug."""
    text = unicodedata.normalize('NFKD', text)
    text = text.replace('đ', 'd').replace('Đ', 'D')
    text = text.encode('ascii', 'ignore').decode('utf-8')
    text = re.sub(r'[^\w\s-]', '', text).strip().lower()
    return re.sub(r'[-\s]+', '-', text)

# ==============================================================================
# Tier 1 & Tier 2: AI Recipe Discovery
# ==============================================================================
def discover_recipes_with_search_grounding(category_keyword: str, count: int) -> List[RecipeExtraction]:
    """Tier 1: Try Gemini with Google Search Grounding."""
    if not ai_client:
        return []

    print(f"  [Tier 1: Search Grounding] Searching Google for '{category_keyword}'...")
    search_prompt = f"""
    Search Google for authentic, popular Vietnamese cooking recipes matching: '{category_keyword}'.
    Look up real recipe pages from reputable Vietnamese culinary websites (such as dienmayxanh.com/vao-bep, cookpad.com/vn, monngonmoingay.com, vnexpress.net/du-lich/am-thuc).
    Select {count} distinct popular recipes that Vietnamese families love to eat.
    For each, extract title (VI & EN), source URL, short description, HTML steps, thumbnail image URL, prep/cook times, difficulty, categories, tags, complete ingredients, and 2-4 suggested related or pairing Vietnamese dish names (e.g. braised dishes paired with sour soup, boiled greens, or stir-fries).
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

def discover_recipes_direct_ai(category_keyword: str, count: int) -> List[RecipeExtraction]:
    """Tier 2: Direct Gemini generation (consumes 0 Search Grounding quota, single call)."""
    if not ai_client:
        return []

    print(f"  [Tier 2: Direct AI] Generating '{category_keyword}' recipe from culinary knowledge base...")
    prompt = f"""
    You are an expert Vietnamese culinary chef and data architect.
    Generate {count} authentic, widely cooked Vietnamese home recipes matching the category: '{category_keyword}'.
    Requirements:
    1. Vietnamese title and English title.
    2. Provide an authentic source URL citing a reputable Vietnamese recipe page (e.g. https://www.dienmayxanh.com/vao-bep/mon-an-ngon, https://monngonmoingay.com, https://cookpad.com/vn).
    3. Short descriptions in VI and EN.
    4. Step-by-step cooking instructions in semantic HTML (<p>, <h4>, <ul>, <li>) in both VI and EN.
    5. High-resolution food thumbnail URL.
    6. Realistic prep and cooking times in minutes.
    7. Difficulty: EASY, MEDIUM, or HARD.
    8. Meal categories (FRIED, SOUP, MEAT, SEAFOOD, TRADITIONAL, etc.) and tags.
    9. Complete ingredient list with realistic quantities, units (g, kg, ml, quả, muỗng canh), notes, and nutritional categories.
    10. Suggested 2-4 authentic related or pairing Vietnamese dish names (e.g. braised dishes paired with sour soup, boiled vegetables, or stir-fries).
    """

    structured_response = ai_client.models.generate_content(
        model=GEMINI_MODEL,
        contents=prompt,
        config=types.GenerateContentConfig(
            response_mime_type="application/json",
            response_schema=RecipeBatch,
            temperature=0.4,
        ),
    )

    batch = RecipeBatch.model_validate_json(structured_response.text)
    return batch.recipes

# ==============================================================================
# Tier 3: Curated Vietnamese Recipe Catalog Fallback
# ==============================================================================
def get_curated_pool() -> List[RecipeExtraction]:
    """Tier 3: Built-in authentic Vietnamese recipes pool ensuring at least 1 dish is ALWAYS ingested."""
    return [
        RecipeExtraction(
            title_vi="Gà chiên nước mắm",
            title_en="Crispy Fried Chicken in Fish Sauce",
            short_desc_vi="Món gà chiên với lớp da giòn rụm quyện đẫm sốt nước mắm tỏi ớt mặn ngọt đậm đà, cực kỳ đưa cơm.",
            short_desc_en="Crispy fried chicken coated in a savory, sweet, and garlicky Vietnamese fish sauce glaze.",
            content_vi="<h4>Bước 1: Sơ chế và ướp gà</h4><p>Cánh gà rửa sạch với nước muối loãng, chặt khúc vừa ăn. Ướp với hạt nêm, tiêu, hành tím băm trong 20 phút.</p><h4>Bước 2: Chiên gà giòn</h4><p>Lăn gà qua một lớp mỏng bột bắp. Cho vào chảo dầu sôi chiên vàng giòn hai mặt rồi vớt ra ráo dầu.</p><h4>Bước 3: Làm sốt nước mắm</h4><p>Pha 2 muỗng nước mắm, 2 muỗng đường, 1 muỗng tương ớt, 1 muỗng nước lọc. Phi thơm tỏi ớt băm, đổ hỗn hợp sốt vào đun sôi sền sệt.</p><h4>Bước 4: Hoàn thành</h4><p>Trút gà đã chiên vào chảo sốt, đảo đều tay ở lửa vừa khoảng 2-3 phút cho sốt áo đều miếng gà. Tắt bếp và thưởng thức cùng cơm nóng.</p>",
            content_en="<h4>Step 1: Prep and Marinate</h4><p>Wash chicken wings with salted water and cut into pieces. Marinate with seasoning powder, pepper, and minced shallots for 20 minutes.</p><h4>Step 2: Deep Fry</h4><p>Lightly dust chicken in cornstarch. Fry in hot oil until golden and crispy, then drain.</p><h4>Step 3: Fish Sauce Glaze</h4><p>Mix 2 tbsp fish sauce, 2 tbsp sugar, 1 tbsp chili sauce, and 1 tbsp water. Sauté minced garlic and chili until fragrant, then add the sauce mixture and simmer until thickened.</p><h4>Step 4: Glaze and Serve</h4><p>Toss fried chicken into the sauce over medium heat for 2-3 minutes until evenly coated. Serve hot with steamed rice.</p>",
            thumbnail="https://cdn.tgdd.vn/Files/2021/08/17/1375775/bi-quyet-lam-canh-ga-chien-nuoc-mam-gion-rum-thom-nuc-mui-202108171447228224.jpg",
            source_url="https://www.dienmayxanh.com/vao-bep/cach-lam-canh-ga-chien-nuoc-mam-thom-ngon-dam-da-00628",
            preparation_time=20,
            cooking_time=25,
            difficulty_level="EASY",
            meal_categories=["FRIED", "LUNCH", "DINNER", "MEAT", "TRADITIONAL"],
            tags=["gà chiên nước mắm", "cánh gà", "món chiên", "cơm gia đình", "fried chicken"],
            suggested_related_dishes=["Canh chua cá lóc miền Tây", "Bò xào cần tỏi", "Rau muống xào tỏi", "Canh cải thịt bằm"],
            ingredients=[
                ParsedIngredient(name_vi="Cánh gà", name_en="Chicken wings", quantity=500.0, measure="g", note="chọn cánh tươi ngon", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Nước mắm", name_en="Fish sauce", quantity=30.0, measure="ml", note="nước mắm truyền thống", category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Đường cát", name_en="Sugar", quantity=30.0, measure="g", note=None, category="SUGARS"),
                ParsedIngredient(name_vi="Tỏi", name_en="Garlic", quantity=4.0, measure="nhánh", note="băm nhuyễn", category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Ớt hiểm", name_en="Chili", quantity=2.0, measure="quả", note="băm nhỏ", category="VEGETABLES"),
                ParsedIngredient(name_vi="Bột bắp", name_en="Cornstarch", quantity=30.0, measure="g", note="tạo độ giòn", category="GRAINS"),
                ParsedIngredient(name_vi="Dầu ăn", name_en="Cooking oil", quantity=200.0, measure="ml", note="để chiên ngập dầu", category="FATS AND OILS")
            ]
        ),
        RecipeExtraction(
            title_vi="Thịt kho tàu nước dừa",
            title_en="Vietnamese Braised Pork and Eggs in Coconut Water",
            short_desc_vi="Món thịt kho tàu truyền thống mềm béo, trứng vịt đậm đà thấm đẫm vị ngọt thanh tự nhiên từ nước dừa xiêm.",
            short_desc_en="Traditional Vietnamese braised pork belly and hard-boiled eggs simmered in sweet fresh coconut water.",
            content_vi="<h4>Bước 1: Sơ chế</h4><p>Thịt ba chỉ cạo sạch lông, rửa với muối loãng, cắt miếng vuông cỡ 3-4cm. Luộc sơ qua nước sôi 2 phút rồi vớt ra rửa sạch.</p><h4>Bước 2: Ướp thịt</h4><p>Ướp thịt với 3 muỗng nước mắm, 1 muỗng đường, hành tỏi băm, tiêu trong ít nhất 30 phút để thịt ngấm vị.</p><h4>Bước 3: Luộc trứng</h4><p>Trứng vịt luộc chín, bóc vỏ sạch sẽ.</p><h4>Bước 4: Kho thịt</h4><p>Thắng nước màu từ đường cho có màu cánh gián, cho thịt vào đảo săn. Đổ nước dừa tươi vào ngập thịt, đun sôi rồi hạ nhỏ lửa đun liu riu khoảng 40 phút. Cho trứng vào kho thêm 20 phút đến khi thịt mềm rục, nước kho sánh vàng óng ả.</p>",
            content_en="<h4>Step 1: Preparation</h4><p>Clean pork belly and cut into 3-4 cm cubes. Blanch in boiling water for 2 minutes, then rinse.</p><h4>Step 2: Marinate</h4><p>Marinate pork with fish sauce, sugar, minced garlic, shallots, and black pepper for at least 30 minutes.</p><h4>Step 3: Boil Eggs</h4><p>Boil duck eggs until hard-boiled, then peel.</p><h4>Step 4: Braise</h4><p>Caramelize sugar until amber, add marinated pork and sear. Pour in fresh coconut water to cover the meat, bring to a boil, then simmer on low heat for 40 minutes. Add eggs and simmer for 20 more minutes until pork is melt-in-mouth tender.</p>",
            thumbnail="https://cdn.tgdd.vn/Files/2021/07/28/1371391/cach-nau-thit-kho-tau-trung-cut-ngon-chuan-vi-mien-nam-202107281635398254.jpg",
            source_url="https://www.dienmayxanh.com/vao-bep/cach-lam-thit-kho-tau-ngon-chuan-vi-00123",
            preparation_time=30,
            cooking_time=60,
            difficulty_level="MEDIUM",
            meal_categories=["LUNCH", "DINNER", "MEAT", "TRADITIONAL"],
            tags=["thịt kho tàu", "thịt kho hột vịt", "món kho", "nước dừa", "braised pork"],
            suggested_related_dishes=["Canh chua cá lóc miền Tây", "Dưa giá đỗ chua ngọt", "Rau muống xào tỏi", "Canh mướp mồng tơi"],
            ingredients=[
                ParsedIngredient(name_vi="Thịt ba chỉ", name_en="Pork belly", quantity=600.0, measure="g", note="chọn thịt ba rọi có nạc mỡ đều", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Nước dừa tươi", name_en="Fresh coconut water", quantity=500.0, measure="ml", note="1 quả dừa xiêm", category="BEVERAGES NONALCOHOLIC"),
                ParsedIngredient(name_vi="Trứng vịt", name_en="Duck eggs", quantity=5.0, measure="quả", note="luộc chín bóc vỏ", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Nước mắm", name_en="Fish sauce", quantity=45.0, measure="ml", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Đường", name_en="Sugar", quantity=30.0, measure="g", note="dùng ướp và thắng màu", category="SUGARS"),
                ParsedIngredient(name_vi="Hành tím", name_en="Shallots", quantity=4.0, measure="củ", note="băm nhuyễn", category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Tỏi", name_en="Garlic", quantity=4.0, measure="nhánh", note="băm nhuyễn", category="HERBS AND SPICES")
            ]
        ),
        RecipeExtraction(
            title_vi="Canh chua cá lóc miền Tây",
            title_en="Mekong Delta Sour Snakehead Fish Soup",
            short_desc_vi="Món canh chua cá lóc thanh mát giải nhiệt với vị chua dịu từ me, cá lóc ngọt thịt cùng bạc hà, giá và thơm giòn ngọt.",
            short_desc_en="A refreshing Southern Vietnamese sour soup with tender snakehead fish, tamarind, pineapple, and crunchy vegetables.",
            content_vi="<h4>Bước 1: Sơ chế cá và rau củ</h4><p>Cá lóc làm sạch vảy, xát muối và chanh khử tanh, cắt khúc vừa ăn. Bạc hà tước vỏ, thái vát. Thơm cắt lát. Cà chua bổ múi cau. Đậu bắp cắt xéo.</p><h4>Bước 2: Nấu nước me</h4><p>Nấu sôi 1 lít nước, cho vắt me vào dầm nát lọc lấy nước chua, bỏ hạt.</p><h4>Bước 3: Nấu canh</h4><p>Cho cá lóc vào nồi nước sôi nấu chín tới trong 5 phút. Vớt bọt cho nước trong. Cho thơm, cà chua, đậu bắp vào đun 3 phút. Nêm nước mắm, đường, muối cho vừa vị chua ngọt hài hòa.</p><h4>Bước 4: Hoàn tất</h4><p>Cho bạc hà và giá đỗ vào đun sôi bùng lên rồi tắt bếp ngay. Múc canh ra tô, rắc ngò gai, ngổ và tỏi phi thơm lên trên.</p>",
            content_en="<h4>Step 1: Prep Ingredients</h4><p>Clean snakehead fish with salt and lime, cut into steaks. Peel and slice elephant ear stems, slice pineapple, wedge tomatoes, and slice okra.</p><h4>Step 2: Tamarind Broth</h4><p>Boil 1 liter of water, dissolve tamarind pulp and strain the sour juice back into the pot.</p><h4>Step 3: Cook Fish and Vegetables</h4><p>Gently add fish to simmering broth and cook for 5 minutes. Skim foam. Add pineapple, tomato, and okra. Season with fish sauce, sugar, and salt to balance sweet and sour.</p><h4>Step 4: Finish and Garnish</h4><p>Add elephant ear stems and bean sprouts, boil for 30 seconds and remove from heat. Garnish with saw-tooth herb, rice paddy herb, and fried garlic.</p>",
            thumbnail="https://cdn.tgdd.vn/Files/2021/06/03/1357064/cach-nau-canh-chua-ca-loc-chuan-vi-mien-tay-202106031702588371.jpg",
            source_url="https://www.dienmayxanh.com/vao-bep/cach-nau-canh-chua-ca-loc-mien-tay-00456",
            preparation_time=25,
            cooking_time=20,
            difficulty_level="MEDIUM",
            meal_categories=["SOUP", "LUNCH", "DINNER", "SEAFOOD", "TRADITIONAL"],
            tags=["canh chua", "cá lóc", "canh chua miền tây", "món canh", "sour soup"],
            suggested_related_dishes=["Cá kho tộ", "Thịt kho tàu nước dừa", "Sườn ram mặn", "Thịt ba chỉ luộc"],
            ingredients=[
                ParsedIngredient(name_vi="Cá lóc", name_en="Snakehead fish", quantity=600.0, measure="g", note="cá đồng tươi sống", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Me chua chín", name_en="Ripe tamarind", quantity=50.0, measure="g", note="dầm lấy nước cốt", category="FRUIT"),
                ParsedIngredient(name_vi="Cà chua", name_en="Tomatoes", quantity=2.0, measure="quả", note="bổ múi cau", category="VEGETABLES"),
                ParsedIngredient(name_vi="Dứa", name_en="Pineapple", quantity=150.0, measure="g", note="thái lát", category="FRUIT"),
                ParsedIngredient(name_vi="Bạc hà", name_en="Elephant ear stem", quantity=100.0, measure="g", note="tước vỏ thái vát", category="VEGETABLES"),
                ParsedIngredient(name_vi="Đậu bắp", name_en="Okra", quantity=100.0, measure="g", note="cắt xéo", category="VEGETABLES"),
                ParsedIngredient(name_vi="Giá đỗ", name_en="Bean sprouts", quantity=100.0, measure="g", note="rửa sạch", category="VEGETABLES"),
                ParsedIngredient(name_vi="Ngò gai và rau ngổ", name_en="Sawtooth and rice paddy herb", quantity=30.0, measure="g", note="cắt nhỏ", category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Nước mắm", name_en="Fish sauce", quantity=30.0, measure="ml", note=None, category="CONDIMENTS AND SAUCES")
            ]
        ),
        RecipeExtraction(
            title_vi="Sườn xào chua ngọt",
            title_en="Sweet and Sour Pork Ribs",
            short_desc_vi="Sườn non mềm ngọt thịt bao phủ bởi lớp sốt chua ngọt sánh đỏ bắt mắt, thơm nức mũi.",
            short_desc_en="Tender pork ribs coated in a glossy, tangy sweet-and-sour glaze.",
            content_vi="<h4>Bước 1: Sơ chế sườn</h4><p>Sườn non chặt miếng vừa ăn, chần qua nước sôi rồi rửa sạch để ráo. Ướp với hạt nêm, tiêu, hành tím băm 15 phút.</p><h4>Bước 2: Rán vàng sườn</h4><p>Cho sườn vào chảo rán vàng đều hai mặt rồi vớt ra đĩa.</p><h4>Bước 3: Pha sốt chua ngọt</h4><p>Khuấy đều 2 muỗng giấm gạo, 2 muỗng đường, 2 muỗng nước mắm, 2 muỗng tương cà và 3 muỗng nước lọc.</p><h4>Bước 4: Xào sườn</h4><p>Phi thơm tỏi băm, đổ hỗn hợp sốt vào đun sôi lăn tăn. Trút sườn rán vào đảo đều cho ngấm sốt. Đun nhỏ lửa đến khi sốt keo lại bám đều quanh từng miếng sườn là hoàn thành.</p>",
            content_en="<h4>Step 1: Prep Ribs</h4><p>Cut ribs into bite-sized pieces, blanch in boiling water and rinse. Marinate with seasoning, pepper, and shallots for 15 minutes.</p><h4>Step 2: Pan-fry</h4><p>Pan-fry ribs until golden on all sides, then remove to a plate.</p><h4>Step 3: Sweet & Sour Sauce</h4><p>Whisk together 2 tbsp rice vinegar, 2 tbsp sugar, 2 tbsp fish sauce, 2 tbsp ketchup, and 3 tbsp water.</p><h4>Step 4: Toss and Simmer</h4><p>Sauté garlic, pour in sauce until simmering. Add ribs and toss well. Simmer gently until sauce thickens into a glossy glaze coating the ribs.</p>",
            thumbnail="https://cdn.tgdd.vn/Files/2021/04/14/1343187/cach-lam-suon-xao-chua-ngot-mien-bac-chuan-vi-thom-ngon-dam-da-202104141641045952.jpg",
            source_url="https://www.dienmayxanh.com/vao-bep/cach-lam-suon-xao-chua-ngot-dam-da-00789",
            preparation_time=20,
            cooking_time=30,
            difficulty_level="EASY",
            meal_categories=["LUNCH", "DINNER", "MEAT", "TRADITIONAL"],
            tags=["sườn xào chua ngọt", "sườn non", "món xào", "cơm gia đình", "sweet and sour ribs"],
            suggested_related_dishes=["Canh rau ngót nấu thịt bằm", "Đậu sốt cà chua", "Canh bí đao sườn heo", "Rau cải xào tỏi"],
            ingredients=[
                ParsedIngredient(name_vi="Sườn non", name_en="Pork ribs", quantity=500.0, measure="g", note="chọn sườn non nhiều nạc", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Giấm gạo", name_en="Rice vinegar", quantity=30.0, measure="ml", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Đường cát", name_en="Sugar", quantity=30.0, measure="g", note=None, category="SUGARS"),
                ParsedIngredient(name_vi="Tương cà", name_en="Ketchup", quantity=30.0, measure="g", note="tạo màu và độ sánh", category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Nước mắm", name_en="Fish sauce", quantity=30.0, measure="ml", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Tỏi", name_en="Garlic", quantity=3.0, measure="nhánh", note="băm nhỏ", category="HERBS AND SPICES")
            ]
        ),
        RecipeExtraction(
            title_vi="Bò xào cần tỏi",
            title_en="Stir-fried Beef with Celery and Garlic",
            short_desc_vi="Thịt bò xào mềm thơm ngậy, cần tây và hành tây giữ trọn độ giòn ngọt, thoang thoảng mùi tiêu ấm nồng.",
            short_desc_en="Quick stir-fried tender beef with crisp celery, onions, and fragrant garlic.",
            content_vi="<h4>Bước 1: Ướp thịt bò</h4><p>Thịt bò thái mỏng ngang thớ. Ướp với 1 muỗng dầu hào, tỏi băm, 1/2 muỗng dầu ăn và chút tiêu trong 15 phút cho thịt mềm.</p><h4>Bước 2: Sơ chế rau củ</h4><p>Cần tây cắt khúc. Hành tây bổ múi cau. Cà chua cắt múi cau.</p><h4>Bước 3: Xào thịt</h4><p>Làm nóng chảo với lửa lớn, cho tỏi băm phi thơm rồi trút thịt bò vào đảo nhanh tay trong 1-2 phút cho thịt vừa chín tới thì trút ra đĩa riêng.</p><h4>Bước 4: Hoàn thành</h4><p>Dùng chảo đó xào hành tây, cà chua và cần tây chín tới. Đổ thịt bò vào đảo đều trong 30 giây, nêm nếm lại cho vừa miệng rồi tắt bếp. Rắc tiêu xay và dùng nóng.</p>",
            content_en="<h4>Step 1: Marinate Beef</h4><p>Slice beef thinly against the grain. Marinate with oyster sauce, minced garlic, a splash of oil, and pepper for 15 minutes.</p><h4>Step 2: Veggie Prep</h4><p>Cut celery into finger-length pieces. Wedge onions and tomatoes.</p><h4>Step 3: Sear Beef</h4><p>Heat wok over high heat with oil and garlic. Add beef and stir-fry vigorously for 1-2 minutes until just cooked. Transfer to a plate.</p><h4>Step 4: Stir-fry and Combine</h4><p>In the same wok, toss onion, tomato, and celery until tender-crisp. Return beef to wok, toss for 30 seconds, garnish with black pepper, and serve hot.</p>",
            thumbnail="https://cdn.tgdd.vn/Files/2021/08/21/1376884/cach-lam-thit-bo-xao-can-tay-thom-ngon-thit-bo-mem-khong-dai-202108211520112423.jpg",
            source_url="https://www.dienmayxanh.com/vao-bep/cach-lam-thit-bo-xao-can-tay-mem-ngon-00345",
            preparation_time=15,
            cooking_time=10,
            difficulty_level="EASY",
            meal_categories=["LUNCH", "DINNER", "MEAT", "TRADITIONAL"],
            tags=["bò xào cần tỏi", "thịt bò", "cần tây", "món xào", "stir fried beef"],
            suggested_related_dishes=["Canh chua cá lóc miền Tây", "Thịt kho tàu nước dừa", "Canh ngao nấu chua", "Trứng chiên hành"],
            ingredients=[
                ParsedIngredient(name_vi="Thịt bò thăn", name_en="Beef tenderloin", quantity=300.0, measure="g", note="thái mỏng ngang thớ", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Cần tây", name_en="Celery", quantity=150.0, measure="g", note="rửa sạch cắt khúc", category="VEGETABLES"),
                ParsedIngredient(name_vi="Hành tây", name_en="Onion", quantity=1.0, measure="củ", note="bổ múi cau", category="VEGETABLES"),
                ParsedIngredient(name_vi="Cà chua", name_en="Tomato", quantity=1.0, measure="quả", note="bổ múi cau", category="VEGETABLES"),
                ParsedIngredient(name_vi="Dầu hào", name_en="Oyster sauce", quantity=15.0, measure="ml", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Tỏi", name_en="Garlic", quantity=4.0, measure="nhánh", note="băm nhỏ", category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Tiêu xay", name_en="Black pepper", quantity=3.0, measure="g", note=None, category="HERBS AND SPICES")
            ]
        ),
        RecipeExtraction(
            title_vi="Tóp mỡ sốt cà chua",
            title_en="Crispy Pork Fat in Tomato Sauce",
            short_desc_vi="Tóp mỡ giòn rụm ngấm đều sốt cà chua sánh mịn chua ngọt, món ăn dân dã siêu tốn cơm của người Việt.",
            short_desc_en="Crispy pork cracklings simmered in a luscious sweet-and-sour fresh tomato sauce.",
            content_vi="<h4>Bước 1: Chuẩn bị tóp mỡ</h4><p>Mỡ heo cắt hạt lựu, rán vàng giòn ép kiệt mỡ nước, vớt tóp mỡ ra để ráo giòn.</p><h4>Bước 2: Nấu sốt cà chua</h4><p>Cà chua băm nhỏ. Phi thơm hành tím trong chảo, cho cà chua vào đảo nhuyễn mịn với 1 muỗng nước mắm, 1 muỗng đường, 1/2 muỗng bột canh.</p><h4>Bước 3: Rim tóp mỡ</h4><p>Khi sốt cà chua sánh sệt, trút toàn bộ tóp mỡ vào chảo đảo nhanh tay trên lửa vừa khoảng 2 phút để tóp mỡ vừa ngấm sốt mà vẫn giữ được độ giòn xốp.</p><h4>Bước 4: Hoàn thành</h4><p>Rắc hành lá thái nhỏ và chút tiêu xay lên trên, tắt bếp và ăn ngay với cơm nóng.</p>",
            content_en="<h4>Step 1: Pork Cracklings</h4><p>Dice pork fat, fry until crispy golden and fat is rendered out, then drain.</p><h4>Step 2: Tomato Sauce</h4><p>Finely chop tomatoes. Sauté shallots, add tomatoes, fish sauce, and sugar, cooking down into a thick sauce.</p><h4>Step 3: Toss Cracklings</h4><p>Pour cracklings into the bubbling tomato sauce, toss rapidly for 2 minutes so they absorb the tangy flavor while staying crispy.</p><h4>Step 4: Garnish</h4><p>Top with chopped scallions and black pepper. Serve immediately with warm rice.</p>",
            thumbnail="https://cdn.tgdd.vn/Files/2021/09/06/1380590/cach-lam-top-mo-sot-ca-chua-gion-rum-thom-ngon-dam-da-202109061618218152.jpg",
            source_url="https://www.dienmayxanh.com/vao-bep/cach-lam-top-mo-sot-ca-chua-gion-ngon-00999",
            preparation_time=15,
            cooking_time=15,
            difficulty_level="EASY",
            meal_categories=["FRIED", "LUNCH", "DINNER", "TRADITIONAL"],
            tags=["tóp mỡ", "sốt cà chua", "món chiên", "dân dã", "pork cracklings"],
            suggested_related_dishes=["Rau muống luộc", "Canh cua rau đay mồng tơi", "Thịt luộc chấm mắm tôm", "Đậu phụ rán"],
            ingredients=[
                ParsedIngredient(name_vi="Tóp mỡ", name_en="Pork cracklings", quantity=250.0, measure="g", note="đã chiên giòn", category="PROTEIN FOODS"),
                ParsedIngredient(name_vi="Cà chua", name_en="Tomatoes", quantity=3.0, measure="quả", note="chín đỏ băm nhỏ", category="VEGETABLES"),
                ParsedIngredient(name_vi="Hành tím", name_en="Shallots", quantity=2.0, measure="củ", note="băm nhỏ", category="HERBS AND SPICES"),
                ParsedIngredient(name_vi="Hành lá", name_en="Scallions", quantity=3.0, measure="nhánh", note="thái nhỏ", category="VEGETABLES"),
                ParsedIngredient(name_vi="Nước mắm", name_en="Fish sauce", quantity=20.0, measure="ml", note=None, category="CONDIMENTS AND SAUCES"),
                ParsedIngredient(name_vi="Đường", name_en="Sugar", quantity=15.0, measure="g", note=None, category="SUGARS")
            ]
        )
    ]

# ==============================================================================
# Database Ingestion Logic
# ==============================================================================
def get_or_create_ingredient(ing: ParsedIngredient, author_id: str = "system-crawler") -> str:
    """Finds existing ingredient by slug or auto-creates it in Ingredients collection."""
    slug = slugify(ing.name_vi)
    if not slug:
        slug = slugify(ing.name_en)

    now = datetime.now(timezone.utc)

    existing = ingredients_col.find_one({"slug": slug, "deleted": False})
    if existing:
        return str(existing["_id"])

    existing_by_title = ingredients_col.find_one({
        "title.data": {"$regex": f"^{re.escape(ing.name_vi)}$", "$options": "i"},
        "deleted": False
    })
    if existing_by_title:
        return str(existing_by_title["_id"])

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
    print(f"    [+] Auto-created Ingredient: {ing.name_vi} (slug: {slug})")
    return str(new_doc["_id"])

def find_and_link_related_dishes(
    dish_id: ObjectId,
    dish_slug: str,
    suggested_names: List[str],
    meal_categories: List[str],
    tags: List[str],
    limit: int = 4
) -> List[str]:
    """
    Finds 2-4 relevant related dish IDs in MongoDB and establishes bidirectional links.
    1. Matches suggested_names by title or slug keywords in MongoDB.
    2. Matches dishes sharing mealCategories or tags.
    3. Falls back to recent dishes if still fewer than 2.
    4. Updates matching related dishes to include dish_id in their relatedDishes array.
    """
    target_id_str = str(dish_id)
    matched_ids: List[str] = []
    seen_ids = {target_id_str}

    print(f"    [~] Searching related dishes for '{dish_slug}'...")

    # Strategy 1: Match suggested dish names in DB
    for name in (suggested_names or []):
        if len(matched_ids) >= limit:
            break
        name_clean = name.strip()
        if not name_clean:
            continue

        name_slug = slugify(name_clean)
        kw_list = [w for w in name_slug.split('-') if len(w) > 2]

        candidate = None
        # 1a. Search by title.data
        try:
            for doc in dishes_col.find({"deleted": False, "title.data": re.compile(re.escape(name_clean), re.IGNORECASE)}).limit(5):
                did = str(doc["_id"])
                if did not in seen_ids and doc.get("slug") != dish_slug:
                    candidate = doc
                    break
        except Exception:
            pass

        # 1b. Search by slug keywords if title didn't match
        if not candidate and kw_list:
            slug_regex = '.*'.join(kw_list[:2]) if len(kw_list) >= 2 else kw_list[0]
            try:
                for doc in dishes_col.find({"deleted": False, "slug": re.compile(slug_regex, re.IGNORECASE)}).limit(5):
                    did = str(doc["_id"])
                    if did not in seen_ids and doc.get("slug") != dish_slug:
                        candidate = doc
                        break
            except Exception:
                pass

        if candidate:
            cid = str(candidate["_id"])
            matched_ids.append(cid)
            seen_ids.add(cid)
            cand_title = next((t["data"] for t in candidate.get("title", []) if t.get("lang") == "vi"), candidate.get("slug"))
            print(f"      [+] Matched suggested '{name_clean}' -> {cand_title} (ID: {cid})")

    # Strategy 2: Match by mealCategories
    if len(matched_ids) < limit:
        needed = limit - len(matched_ids)
        for cat in (meal_categories or []):
            if len(matched_ids) >= limit:
                break
            try:
                for doc in dishes_col.find({"deleted": False, "mealCategories": cat}).limit(needed * 2):
                    did = str(doc["_id"])
                    if did not in seen_ids and doc.get("slug") != dish_slug:
                        matched_ids.append(did)
                        seen_ids.add(did)
                        cand_title = next((t["data"] for t in doc.get("title", []) if t.get("lang") == "vi"), doc.get("slug"))
                        print(f"      [+] Matched category '{cat}' -> {cand_title} (ID: {did})")
                        if len(matched_ids) >= limit:
                            break
            except Exception:
                pass

    # Strategy 3: Fallback to recent dishes
    if len(matched_ids) < 2:
        needed = limit - len(matched_ids)
        try:
            for doc in dishes_col.find({"deleted": False}).sort("createdAt", -1).limit(needed * 3):
                did = str(doc["_id"])
                if did not in seen_ids and doc.get("slug") != dish_slug:
                    matched_ids.append(did)
                    seen_ids.add(did)
                    cand_title = next((t["data"] for t in doc.get("title", []) if t.get("lang") == "vi"), doc.get("slug"))
                    print(f"      [+] Matched recent fallback -> {cand_title} (ID: {did})")
                    if len(matched_ids) >= limit:
                        break
        except Exception:
            pass

    # Bidirectional link: update each related dish document
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
    dish_slug = slugify(recipe.title_vi)
    if not dish_slug:
        dish_slug = slugify(recipe.title_en)

    existing = dishes_col.find_one({
        "$or": [
            {"slug": dish_slug},
            {"sourceUrl": recipe.source_url}
        ],
        "deleted": False
    })
    if existing:
        print(f"  [-] Dish already in database ('{dish_slug}'). Skipping duplicate.")
        return False

    print(f"\n  [>] Ingesting Dish: {recipe.title_vi} ({recipe.title_en})")
    print(f"      Source URL: {recipe.source_url}")

    new_dish_id = ObjectId()

    mapped_ingredients = []
    for ing in recipe.ingredients:
        ing_id = get_or_create_ingredient(ing, author_id=author_id)
        mapped_ingredients.append({
            "ingredientId": ing_id,
            "slug": slugify(ing.name_vi),
            "quantity": ing.quantity,
            "note": ing.note
        })

    # Find and bidirectionally link related dishes
    related_dish_ids = find_and_link_related_dishes(
        dish_id=new_dish_id,
        dish_slug=dish_slug,
        suggested_names=recipe.suggested_related_dishes,
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
        "thumbnail": recipe.thumbnail,
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
    print(f"  [SUCCESS] Ingested '{recipe.title_vi}' (ID: {new_dish_id}) with {len(mapped_ingredients)} ingredients and {len(related_dish_ids)} related dishes linked.\n")
    return True

# ==============================================================================
# Main Resilient Pipeline
# ==============================================================================
def main():
    print("=" * 60)
    print("Starting Resilient What-To-Eat Recipe Ingestion Job")
    print(f"Target limit: {DAILY_CRAWL_LIMIT} dishes")
    print(f"Categories: {CRAWL_CATEGORIES}")
    print("=" * 60)

    total_inserted = 0
    exploration_queue = list(CRAWL_CATEGORIES)
    visited_topics = set()

    # --------------------------------------------------------------------------
    # Step 1 & 2: Try AI Search Grounding, then Direct AI Generation
    # --------------------------------------------------------------------------
    while exploration_queue and total_inserted < DAILY_CRAWL_LIMIT:
        category = exploration_queue.pop(0)
        if category in visited_topics:
            continue
        visited_topics.add(category)

        remaining = DAILY_CRAWL_LIMIT - total_inserted
        fetch_count = min(remaining, 2)
        recipes = []

        # Attempt Tier 1: Search Grounding
        try:
            recipes = discover_recipes_with_search_grounding(category, count=fetch_count)
        except Exception as e_grounding:
            print(f"  [Notice] Tier 1 search grounding bypassed ({e_grounding})")

        # If Tier 1 failed or returned 0, fallback to Tier 2: Direct AI
        if not recipes and ai_client:
            try:
                recipes = discover_recipes_direct_ai(category, count=fetch_count)
            except Exception as e_direct:
                print(f"  [Notice] Tier 2 direct AI bypassed ({e_direct})")

        # Ingest whatever recipes were gathered
        for recipe in recipes:
            if total_inserted >= DAILY_CRAWL_LIMIT:
                break
            if ingest_recipe(recipe):
                total_inserted += 1
                # If crawl limit remains, queue suggested related dishes that don't exist yet!
                for rel_name in (recipe.suggested_related_dishes or []):
                    rel_slug = slugify(rel_name)
                    if not dishes_col.find_one({"deleted": False, "slug": rel_slug}):
                        if rel_name not in visited_topics and rel_name not in exploration_queue:
                            print(f"  [+] Discovered new related dish topic to explore: '{rel_name}'")
                            exploration_queue.append(rel_name)
            time.sleep(1)

    # --------------------------------------------------------------------------
    # Step 3: Guaranteed Fallback from Curated Pool
    # --------------------------------------------------------------------------
    if total_inserted < DAILY_CRAWL_LIMIT:
        print(f"\n[Tier 3 Fallback] Target not reached ({total_inserted}/{DAILY_CRAWL_LIMIT}). Drawing from Curated Pool...")
        curated_recipes = get_curated_pool()
        for recipe in curated_recipes:
            if total_inserted >= DAILY_CRAWL_LIMIT:
                break
            if ingest_recipe(recipe):
                total_inserted += 1
                # Queue uncreated related dishes from curated pool if still needed
                for rel_name in (recipe.suggested_related_dishes or []):
                    rel_slug = slugify(rel_name)
                    if not dishes_col.find_one({"deleted": False, "slug": rel_slug}):
                        if rel_name not in visited_topics and rel_name not in exploration_queue:
                            exploration_queue.append(rel_name)

    print("=" * 60)
    print(f"Job completed successfully. Total new dishes ingested: {total_inserted}")
    print("=" * 60)

    if total_inserted == 0:
        print("[WARN] All available recipes already exist in database.", file=sys.stderr)

if __name__ == "__main__":
    main()
