# Automated Recipe Discovery & AI Ingestion Pipeline

This service automatically discovers authentic Vietnamese recipes from the web (including Điện Máy Xanh, Cookpad, Món Ngon Mỗi Ngày, etc.) using **Google Gemini 2.5 Flash** with **Google Search Grounding**, extracts and translates them into bilingual structure, records the original source URL, resolves/creates ingredients, and ingests them into MongoDB.

## Features
- **Zero manual editing**: Translates Vietnamese and English titles, formats step-by-step semantic HTML, estimates cooking/prep time, tags, difficulty, and meal categories.
- **Source URL tracking**: Saves the original source web URL in `sourceUrl` on each `Dish` document.
- **Ingredient Resolution**: Extracts exact ingredient quantities and units. If an ingredient doesn't exist in MongoDB yet, it is automatically created in the `Ingredients` collection, and linked via `ingredientId`.
- **Duplicate Prevention**: Automatically checks dish and ingredient slugs and `sourceUrl` to prevent duplicate ingestion.
- **Cloud-Friendly**: Uses Google Search Grounding instead of scraping raw web pages directly, avoiding firewall geo-blocks and timeouts on GitHub Actions.
- **Scheduled**: Runs automatically once a day at 08:00 AM VN time (01:00 UTC) via GitHub Actions, or manually on demand.

## Required Environment Variables / GitHub Secrets
- `MONGODB_CONNECTION_STRING`: Connection URI (MongoDB Atlas or self-hosted)
- `DATABASE_NAME`: Database name (e.g. `what-to-eat`)
- `GEMINI_API_KEY`: API Key from Google AI Studio (`aistudio.google.com`)
- `DAILY_CRAWL_LIMIT`: Number of dishes to crawl per run (default: `5`)
- `CRAWL_CATEGORIES`: Comma-separated search topics (default: `món chiên giòn ngon,món xào đưa cơm,món kho đậm đà,món canh thanh mát,món nướng thơm lừng,món hấp thanh đạm`)

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
