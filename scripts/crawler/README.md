# Automated World Recipe Discovery & AI Ingestion Pipeline

This service automatically discovers authentic recipes from cuisines all over the world (including Italian, Japanese, Mexican, Indian, Thai, Korean, French, Chinese, Mediterranean, Middle Eastern, Spanish, American, Vietnamese, Moroccan, Turkish, and Latin American dishes) from reputable international culinary websites (such as Serious Eats, BBC Good Food, Allrecipes, RecipeTin Eats, Just One Cookbook, Bon Appétit, Food Network, and Cookpad) using **Google Gemini 3.8 Flash** with **Google Search Grounding**, with automatic fallback to Direct AI generation and a Curated World Recipe Pool. It extracts and translates recipes into bilingual (`en` & `vi`) structure, records the original source URL, resolves/creates ingredients, bidirectionally links related dishes, and ingests them into MongoDB.

## Features
- **Global cuisine diversity**: Shuffles across 16 world cuisines/categories and avoids duplicating existing database dishes so every run adds varied global dishes.
- **Zero manual editing**: Translates English and Vietnamese titles, formats step-by-step semantic HTML, estimates cooking/prep time, tags (including cuisine/country), difficulty, and meal categories.
- **Source URL tracking**: Saves the original source web URL in `sourceUrl` on each `Dish` document.
- **Ingredient Resolution**: Extracts exact ingredient quantities and units. If an ingredient doesn't exist in MongoDB yet, it is automatically created in the `Ingredients` collection, and linked via `ingredientId`.
- **Duplicate Prevention**: Automatically checks dish and ingredient slugs (both English and Vietnamese) and specific `sourceUrl` paths to prevent duplicate ingestion.
- **Cloud-Friendly**: Uses Google Search Grounding instead of scraping raw web pages directly, avoiding firewall geo-blocks and timeouts on GitHub Actions.
- **Scheduled**: Runs automatically once a day at 03:00 UTC (10:00 AM GMT+7) via GitHub Actions, or manually on demand.

## Required Environment Variables / GitHub Secrets
- `MONGODB_CONNECTION_STRING`: Connection URI (MongoDB Atlas or self-hosted)
- `DATABASE_NAME`: Database name (e.g. `what-to-eat`)
- `GEMINI_API_KEY`: API Key from Google AI Studio (`aistudio.google.com`)
- `DAILY_CRAWL_LIMIT`: Number of dishes to crawl per run (default: `5`)
- `CRAWL_CATEGORIES`: Comma-separated search cuisines/topics (default: `Italian cuisine,Japanese cuisine,Mexican cuisine,Indian cuisine,Thai cuisine,Korean cuisine,French cuisine,Chinese cuisine,Greek and Mediterranean cuisine,Middle Eastern cuisine,Spanish cuisine,American comfort food,Vietnamese cuisine,Moroccan and North African cuisine,Turkish cuisine,Latin American cuisine`)

## Running Locally

1. Create a Python virtual environment:
   ```bash
   python3 -m venv venv
   source venv/bin/activate
   pip install -r requirements.txt
   ```

2. Set your environment variables in `.env`:
   ```bash
   MONGODB_CONNECTION_STRING="mongodb://..."
   DATABASE_NAME="what-to-eat"
   GEMINI_API_KEY="your-gemini-api-key"
   DAILY_CRAWL_LIMIT="3"
   ```

3. Run the script:
   ```bash
   python auto_crawl_ingest.py
   ```
