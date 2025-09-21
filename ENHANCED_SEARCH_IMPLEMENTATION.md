# Enhanced Search Implementation - Google-like Search for Dishes

## Overview

I've enhanced the `FindWithScore` function and added new search capabilities to provide Google-like search functionality for dishes. The implementation includes advanced relevance scoring, fuzzy matching, typo tolerance, and intelligent result ranking.

## Key Features

### 1. Advanced Relevance Scoring
- **Field-weighted scoring**: Different fields have different importance weights
  - Title: Highest priority (100/80/60 points for exact/starts-with/contains)
  - Slug: High priority (90/70/50 points)
  - Short Description: High priority (80/60/40 points)
  - Tags: Medium-high priority (70/50/30 points)
  - Content: Medium priority (60/40/20 points)
  - Ingredients: Medium priority (50/30/15 points)

### 2. Multi-Pattern Matching
- **Exact matches**: Highest scoring for precise matches
- **Starts-with matches**: High scoring for prefix matches
- **Contains matches**: Medium scoring for substring matches
- **Phrase matching**: Special scoring for multi-word queries with proximity

### 3. Smart Query Processing
- **Multi-word handling**: Processes each word individually and as phrases
- **Word proximity**: Rewards documents where query words appear close together
- **Individual word scoring**: Bonus points for matching individual words in multi-word queries

### 4. Fuzzy Search & Typo Tolerance
- **Common cooking term corrections**: Built-in dictionary of common misspellings
  - "chiken" → "chicken"
  - "tomatoe" → "tomato"
  - "spagetti" → "spaghetti"
  - And 25+ more common food-related typos

### 5. Smart Search Strategies
- **Strategy 1**: Try exact phrase search first
- **Strategy 2**: If few results, try individual word matching
- **Strategy 3**: If still few results, apply fuzzy search with typo corrections

### 6. Popularity Boost
- **Recency boost**: Newer dishes (within 3 months) get slight scoring boost
- **Engagement potential**: Framework ready for view count, rating-based scoring

## API Endpoints

### 1. Enhanced Search (Default)
```
GET /api/dish/
```
- Uses `FindSmart()` for keyword searches
- Falls back to regular `Find()` for filter-only queries
- Automatically applies best search strategy

### 2. Scored Search
```
GET /api/dish/search/
```
- Uses `FindWithScore()` directly
- Returns results with relevance scoring
- Best for when you need consistent scoring behavior

### 3. Fuzzy Search
```
GET /api/dish/search/fuzzy/
```
- Uses `FindWithFuzzyScore()`
- Handles typos and misspellings
- Best for handling user input errors

### 4. Suggestions (Unchanged)
```
GET /api/dish/suggestions/
```
- Auto-complete functionality
- Prefix matching for search suggestions

## Query Parameters

All endpoints support the same parameters:
- `keyword`: Search term (enhanced processing)
- `page`: Page number (default: 1)
- `limit`: Results per page (default: 10)
- `tags[]`: Filter by tags
- `preparationTimeFrom/To`: Time range filters
- `cookingTimeFrom/To`: Time range filters
- `difficultLevels[]`: Difficulty filters
- `mealCategories[]`: Meal type filters
- `ingredientCategories[]`: Ingredient category filters
- `ingredients[]`: Specific ingredients
- `labels[]`: Label filters

## Search Examples

### Simple Search
```
GET /api/dish/?keyword=chicken
```
- Finds dishes with "chicken" in title, description, ingredients
- Ranks by relevance score

### Multi-word Search
```
GET /api/dish/?keyword=steamed chicken rice
```
- Looks for exact phrase "steamed chicken rice"
- Also matches individual words with proximity scoring
- Rewards documents with words appearing close together

### Typo Tolerance
```
GET /api/dish/search/fuzzy/?keyword=chiken tomatos
```
- Corrects "chiken" → "chicken"
- Corrects "tomatos" → "tomatoes"
- Returns relevant results despite typos

### Complex Query
```
GET /api/dish/?keyword=spicy noodles&mealCategories[]=dinner&cookingTimeFrom=10&cookingTimeTo=30
```
- Searches for "spicy noodles" with relevance scoring
- Filters by dinner meal category
- Limits to 10-30 minute cooking time

## Technical Implementation

### Database Aggregation Pipeline
The enhanced search uses MongoDB aggregation pipeline with:
1. **Match Stage**: Apply filters (non-deleted, categories, time ranges)
2. **AddFields Stage**: Calculate relevance scores using complex expressions
3. **Match Stage**: Filter out zero-score results
4. **Sort Stage**: Order by score DESC, then creation date DESC
5. **Pagination**: Skip and limit for paging

### Scoring Algorithm
```javascript
Total Score = Title Score + Description Score + Content Score + Slug Score + Tag Score + Ingredient Score + Word Match Bonuses + Recency Boost
```

### Regex Patterns
- **Exact**: `^keyword$` (100% match)
- **Starts With**: `^keyword` (prefix match)
- **Contains**: `keyword` (substring match)
- **Phrase**: `word1.*word2.*word3` (proximity match)

## Performance Considerations

### Optimizations Applied
1. **Early filtering**: Apply non-search filters first to reduce dataset
2. **Efficient regex**: Use `regexp.QuoteMeta()` to escape special characters
3. **Fallback strategy**: Use simpler `Find()` when no keyword provided
4. **Smart limits**: Limit aggregation pipeline stages appropriately

### Recommended Indexes
```javascript
// Compound index for common filters
db.dishes.createIndex({ "deleted": 1, "mealCategories": 1, "createdAt": -1 })

// Text search index for full-text search fallback
db.dishes.createIndex({ 
  "title.data": "text", 
  "shortDescription.data": "text", 
  "content.data": "text", 
  "tags": "text" 
})

// Individual field indexes for specific searches
db.dishes.createIndex({ "slug": 1 })
db.dishes.createIndex({ "tags": 1 })
db.dishes.createIndex({ "ingredients.slug": 1 })
```

## Future Enhancements

### Planned Improvements
1. **Machine Learning**: Add ML-based relevance scoring
2. **User Behavior**: Track clicks/views for popularity scoring
3. **Semantic Search**: Add vector-based similarity matching
4. **Analytics**: Track search patterns for optimization
5. **Caching**: Implement Redis caching for popular searches
6. **A/B Testing**: Framework for testing different scoring algorithms

### Monitoring Recommendations
1. **Search Analytics**: Track popular keywords, zero-result queries
2. **Performance Metrics**: Monitor query execution times
3. **User Engagement**: Track click-through rates on search results
4. **Error Monitoring**: Log failed searches and edge cases

## Testing Examples

You can test the enhanced search with these examples:

```bash
# Basic search
curl "http://localhost:8080/api/dish/?keyword=chicken"

# Multi-word search
curl "http://localhost:8080/api/dish/?keyword=steamed%20chicken%20rice"

# Typo tolerance
curl "http://localhost:8080/api/dish/search/fuzzy/?keyword=chiken"

# Complex search with filters
curl "http://localhost:8080/api/dish/?keyword=spicy&mealCategories[]=dinner"
```

The enhanced search system provides a robust, Google-like experience that handles user queries intelligently while maintaining fast performance and relevant results.
