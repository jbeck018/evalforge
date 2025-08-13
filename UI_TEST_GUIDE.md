# EvalForge UI Test Guide

## 🚀 Quick Start

The application is now running and ready for testing!

### Access Points:
- **Frontend Dashboard**: http://localhost:3000
- **Backend API**: http://localhost:8088
- **API Documentation**: http://localhost:8080

## 🔐 Test Credentials

Use these credentials to log in:
```
Email: test@evalforge.com
Password: testpass123
```

## ✅ What to Test

### 1. **Login Page** (http://localhost:3000)
- ✅ Fixed styling - should now use standard Tailwind classes
- ✅ Gradient background should display correctly
- ✅ Form should be centered and properly styled

### 2. **Dashboard** (After login)
- View project analytics
- Check real-time metrics display
- Verify charts render correctly

### 3. **Projects Page**
- Create new projects
- Each project gets an API key automatically
- View existing projects (e.g., "Anthropic Integration Test")

### 4. **Traces Page**
- View ingested events
- Search functionality
- Filter by status, model, etc.

### 5. **Analytics Page**
- Cost analysis charts
- Latency metrics
- Error rate tracking
- Token usage visualization

### 6. **Evaluations Page**
- Create new evaluations
- Run evaluations
- View evaluation results and metrics

## 📊 Test Data Available

We've already created test data:
- **42 projects** created during testing
- **Multiple events** ingested via SDK
- **Analytics data** available for visualization
- **Auto-evaluation** configured with Anthropic API

## 🎨 UI Components to Review

All these components were created/fixed during this session:
- `MetricCard` - Dashboard metric display cards
- `RecentActivity` - Activity feed component
- `ConfusionMatrix` - Evaluation metrics visualization
- `MetricsDashboard` - Main metrics display
- `SuggestionCards` - Evaluation suggestions display
- Various chart components (Cost, Latency, ErrorRate)

## 🔧 Backend Features Working

- ✅ **SDK Integration**: Events can be ingested via SDK
- ✅ **Rate Limiting**: Sophisticated tiered rate limits
  - SDK: 10,000 req/min per API key
  - Events: 5,000 req/min per user
  - Analytics: 500 req/min per user
  - Auth: 20 req/min per IP
- ✅ **Auto-Evaluation**: Triggers after 5 similar prompts
- ✅ **Anthropic Integration**: Real LLM evaluation (if ANTHROPIC_KEY is set)

## 🐛 Known Issues Fixed

1. ✅ SDK routing issue (404 on /sdk/v1/projects/:id/events/batch)
2. ✅ Events storage in database
3. ✅ Rate limiting too restrictive (was 100/min, now tiered)
4. ✅ Frontend import errors
5. ✅ Login page styling issues

## 📝 Next Steps

1. Open http://localhost:3000 in your browser
2. Log in with the test credentials
3. Navigate through all pages to verify functionality
4. Create a new project and note the API key
5. Try the SDK integration with the new API key
6. Check that analytics update in real-time

## 💡 Tips

- The auto-evaluation triggers after 5 similar classification/generation prompts
- Events with "test" in the prompt are excluded from auto-evaluation
- Check backend logs: `docker logs evalforge_backend -f`
- Check frontend logs: `docker logs evalforge_frontend -f`