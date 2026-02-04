# Context Window Consumption Analysis

## Issues Identified

### 1. 🔴 CRITICAL: Pretty-Printed JSON (2-space indentation)
**Location:** `internal/handlers/base.go:28`
```go
result, err := json.MarshalIndent(response, "", "  ")  // 2-space indent
```

**Impact:** 
- Adds ~30-40% extra characters for whitespace
- Example: `{"id":"123"}` becomes:
```json
{
  "id": "123"
}
```
- For 100 incidents, this adds thousands of unnecessary characters

**Solution:** Use compact JSON (`json.Marshal`) instead of `json.MarshalIndent`

### 2. 🔴 CRITICAL: Verbose Pagination Messages
**Location:** `internal/handlers/incidents.go:237-240`
```go
response["FETCH_NEXT_PAGE"] = map[string]interface{}{
    "action":  "REQUIRED - You must call list_incidents again to get remaining results",
    "after":   resp.PaginationMeta.After,
    "message": fmt.Sprintf("Fetched %d of %d incidents (%.1f%%). Call list_incidents again with after='%s' plus same filters. Repeat until has_more_results=false.", ...),
}
```

**Impact:**
- Adds 200+ characters per response
- Repeated in every paginated response
- Redundant information (action + message say the same thing)

**Solution:** Simplify to just essential data

### 3. 🟡 MEDIUM: Redundant Pagination Data
**Location:** `internal/handlers/incidents.go:231-236`
```go
response["pagination_progress"] = map[string]interface{}{
    "records_fetched":  recordsFetched,
    "total_records":    totalRecords,
    "remaining":        totalRecords - recordsFetched,  // Redundant
    "progress_percent": fmt.Sprintf("%.1f%%", ...),     // Redundant
}
```

**Impact:**
- `remaining` can be calculated: `total - fetched`
- `progress_percent` can be calculated: `(fetched/total)*100`
- Adds ~100 characters per response

**Solution:** Remove calculated fields, keep only raw data

### 4. 🟡 MEDIUM: Full Incident Objects by Default
**Location:** `internal/handlers/incidents.go`

**Impact:**
- Even with `summary=true`, responses can be large
- Each incident has many fields (custom fields, role assignments, etc.)
- 25 incidents × 500 bytes each = 12.5KB per response

**Current Mitigation:** Already implemented summary mode (good!)
**Additional Solution:** Add field selection parameter

### 5. 🟢 LOW: Verbose Error Messages
**Location:** Various handlers

**Impact:**
- Error messages include full response bodies
- Can be very large for API errors

**Solution:** Truncate error responses to first 500 characters

## Estimated Impact

### Current Response Sizes (estimated):
- **list_incidents (25 items, summary mode):** ~15-20KB
  - Pretty-print overhead: ~5KB (33%)
  - Verbose pagination: ~0.5KB (3%)
  - Redundant data: ~0.5KB (3%)
  
- **list_incidents (25 items, full mode):** ~50-100KB
  - Pretty-print overhead: ~15-30KB (30-40%)
  - Each incident: 2-4KB

### After Fixes:
- **list_incidents (25 items, summary mode):** ~10-12KB (-40%)
- **list_incidents (25 items, full mode):** ~35-70KB (-30%)

## Recommendations Priority

### Immediate (High Impact, Low Risk):
1. ✅ **Remove pretty-printing** - Use compact JSON
2. ✅ **Simplify pagination messages** - Remove verbose text
3. ✅ **Remove redundant pagination fields** - Keep only essential data

### Short-term (Medium Impact, Low Risk):
4. ✅ **Add response size limits** - Truncate if > 50KB
5. ✅ **Add field selection** - Let users choose which fields to return
6. ✅ **Truncate error messages** - Limit to 500 chars

### Long-term (Lower Priority):
7. **Streaming responses** - For very large result sets
8. **Compression** - Gzip responses (if MCP supports it)
9. **Response caching** - Cache formatted responses

## Implementation Plan

### Phase 1: Quick Fixes (30 min)
- Switch to compact JSON
- Simplify pagination messages
- Remove redundant fields

### Phase 2: Safety Limits (30 min)
- Add max response size (50KB default)
- Truncate with warning if exceeded
- Add character count to responses

### Phase 3: Advanced (1 hour)
- Field selection parameter
- Response size optimization
- Comprehensive testing