# Enhanced Search Implementation Guide

## Overview
This document outlines the enhanced search capabilities implemented for the dish search functionality to improve upon the basic full-text search.

## Problems with Previous Implementation
1. **Limited Full-Text Search**: MongoDB's `$text` operator provides basic text matching but lacks flexibility
2. **No Relevance Scoring**: Results weren't ranked by relevance
3. **No Fuzzy Matching**: Exact matches only, no partial or fuzzy search
4. **Poor User Experience**: No auto-suggestions or intelligent search assistance

## Enhanced Features

### 1. Improved Find Function (`Find`)
**Endpoint**: `GET /api/dishes/`

**Enhancements**:
- **Multi-field Search**: Searches across title, description, content, slug, tags, and ingredient slugs
- **Case-Insensitive**: Uses regex with `i` flag for case-insensitive matching
- **Fuzzy Matching**: Partial string matching using regex patterns
- **Multi-language Support**: Searches through all language variants in `MultiLanguage` fields

**Usage Example**:
```
GET /api/dishes/?keyword=chicken&page=1&limit=10
```

### 2. Weighted Search Function (`FindWithScore`)
**Endpoint**: `GET /api/dishes/search/`

**Features**:
- **Relevance Scoring**: Different weights for different fields:
  - Title: 10 points (highest priority)
  - Slug: 8 points
  - Short Description: 7 points
  - Tags: 5 points
  - Content: 3 points (lowest priority)
- **Smart Ranking**: Results sorted by relevance score, then by creation date
- **Aggregation Pipeline**: Uses MongoDB aggregation for complex scoring logic

**Usage Example**:
```
GET /api/dishes/search/?keyword=spicy%20chicken&tags=dinner&mealCategories=main-course
```

**Response Format**:
```json
{
  "dishes": [
    {
      "slug": "spicy-chicken-curry",
      "title": [{"lang": "en", "data": "Spicy Chicken Curry"}],
      "searchScore": 18,
      // ... other fields
    }
  ],
  "count": 25
}
```

### 3. Auto-Suggestion Function (`FindSuggestions`)
**Endpoint**: `GET /api/dishes/suggestions/`

**Features**:
- **Real-time Suggestions**: Provides auto-complete suggestions as user types
- **Multi-source**: Suggestions from titles, slugs, and tags
- **Frequency-based**: Most common suggestions appear first
- **Prefix Matching**: Matches terms that start with the input keyword

**Usage Example**:
```
GET /api/dishes/suggestions/?keyword=chick&limit=10
```

**Response Format**:
```json
{
  "suggestions": [
    "chicken curry",
    "chicken tikka",
    "chicken biryani",
    "chicken soup"
  ]
}
```

## Implementation Details

### Search Algorithm Improvements

1. **Regex-based Fuzzy Search**:
   ```go
   regexPattern := primitive.Regex{Pattern: *query.Keyword, Options: "i"}
   ```

2. **Multi-field OR Query**:
   ```go
   searchConditions := bson.A{
       bson.D{{Key: "$text", Value: bson.D{{Key: "$search", Value: query.Keyword}}}},
       bson.D{{Key: "title.data", Value: regexPattern}},
       bson.D{{Key: "shortDescription.data", Value: regexPattern}},
       // ... more fields
   }
   filter = append(filter, bson.E{Key: "$or", Value: searchConditions})
   ```

3. **Weighted Scoring System**:
   ```go
   "searchScore": bson.D{{
       Key: "$add", Value: bson.A{
           // Title match (weight: 10)
           bson.D{{Key: "$cond", Value: bson.A{...}}},
           // Other field matches with different weights
       }
   }}
   ```

### Performance Considerations

1. **Database Indexes**: Ensure these indexes exist for optimal performance:
   ```javascript
   // MongoDB shell commands
   db.dishes.createIndex({"title.data": "text", "shortDescription.data": "text", "content.data": "text"})
   db.dishes.createIndex({"slug": 1})
   db.dishes.createIndex({"tags": 1})
   db.dishes.createIndex({"deleted": 1, "createdAt": -1})
   ```

2. **Pagination**: All functions support pagination to handle large result sets efficiently

3. **Aggregation Optimization**: Uses MongoDB aggregation pipeline for complex operations

## Usage Recommendations

### For Basic Search
Use the enhanced `Find` function (`GET /api/dishes/`) for:
- Simple keyword searches
- When you need basic fuzzy matching
- Backward compatibility with existing implementations

### For Advanced Search
Use `FindWithScore` (`GET /api/dishes/search/`) for:
- When relevance ranking is important
- Complex multi-criteria searches
- When you need the best possible search results

### For Auto-complete
Use `FindSuggestions` (`GET /api/dishes/suggestions/`) for:
- Search input auto-completion
- Improving user experience
- Reducing typos and search errors

## Frontend Integration Examples

### Search with Auto-complete
```typescript
// Auto-suggestions
async getSuggestions(keyword: string) {
  const response = await fetch(`/api/dishes/suggestions/?keyword=${keyword}&limit=10`);
  return response.json();
}

// Enhanced search
async searchDishes(keyword: string, filters: any) {
  const params = new URLSearchParams({
    keyword,
    ...filters,
    page: '1',
    limit: '20'
  });
  const response = await fetch(`/api/dishes/search/?${params}`);
  return response.json();
}
```

### Search Component Example
```typescript
class DishSearchComponent {
  private searchWithScore = true; // Toggle between basic and advanced search
  
  search(keyword: string, filters: any) {
    const endpoint = this.searchWithScore ? '/api/dishes/search/' : '/api/dishes/';
    // ... implementation
  }
}
```

## Future Enhancements

1. **Elasticsearch Integration**: For even more advanced full-text search capabilities
2. **Machine Learning**: Use ML models for better relevance scoring
3. **Personalization**: User-based search result customization
4. **Analytics**: Track search queries and improve based on user behavior
5. **Caching**: Implement Redis caching for frequently searched terms

## Migration Notes

- The original `Find` function is enhanced but maintains backward compatibility
- New endpoints are additive and don't break existing functionality
- Frontend applications can gradually migrate to use the new enhanced search features
- All new functions follow the same authentication and authorization patterns as existing endpoints
