# EvalForge API Documentation

## Base URL
- Local: `http://localhost:8088`
- Production: `https://api.evalforge.ai`

## Authentication

All API endpoints (except health and auth endpoints) require authentication using a Bearer token.

### Headers
```
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json
```

## API Endpoints

### Health Check

#### GET /health
Check if the API is running and healthy.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": 1691234567,
  "version": "1.0.0"
}
```

---

### Authentication

#### POST /api/auth/register
Register a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

#### POST /api/auth/login
Login with existing credentials.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

---

### Projects

#### GET /api/projects
List all projects for the authenticated user.

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20)

**Response (200):**
```json
{
  "projects": [
    {
      "id": 1,
      "name": "My AI App",
      "description": "Production chatbot",
      "api_key": "proj_abc123...",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 10,
  "page": 1,
  "limit": 20
}
```

#### POST /api/projects
Create a new project.

**Request Body:**
```json
{
  "name": "My AI App",
  "description": "Production chatbot application"
}
```

**Response (201):**
```json
{
  "project": {
    "id": 1,
    "name": "My AI App",
    "description": "Production chatbot application",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "api_key": "proj_abc123..."
}
```

#### GET /api/projects/:id
Get project details.

**Response (200):**
```json
{
  "project": {
    "id": 1,
    "name": "My AI App",
    "description": "Production chatbot",
    "settings": {
      "auto_evaluation_enabled": true,
      "evaluation_threshold": 0.8,
      "max_traces_per_day": 10000,
      "retention_days": 30
    },
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### DELETE /api/projects/:id
Delete a project and all associated data.

**Response (200):**
```json
{
  "message": "Project deleted successfully"
}
```

---

### Event Ingestion

#### POST /api/events
Ingest events in batch.

**Request Body:**
```json
{
  "events": [
    {
      "id": "evt_123",
      "project_id": 1,
      "trace_id": "trace_456",
      "span_id": "span_789",
      "operation_type": "chat",
      "start_time": "2024-01-01T00:00:00Z",
      "end_time": "2024-01-01T00:00:01Z",
      "duration_ms": 1000,
      "status": "success",
      "input": {
        "prompt": "What is the weather?"
      },
      "output": {
        "response": "I cannot provide weather information."
      },
      "metadata": {
        "user_id": "user_123",
        "session_id": "session_456"
      },
      "tokens": {
        "prompt": 10,
        "completion": 15,
        "total": 25
      },
      "cost": 0.0005,
      "provider": "openai",
      "model": "gpt-3.5-turbo"
    }
  ]
}
```

**Response (201):**
```json
{
  "message": "Events ingested successfully",
  "count": 1
}
```

#### GET /api/projects/:id/events
Retrieve events for a project.

**Query Parameters:**
- `start_time` (ISO 8601): Filter by start time
- `end_time` (ISO 8601): Filter by end time
- `operation_type` (string): Filter by operation type
- `status` (string): Filter by status
- `limit` (int): Number of events (default: 100)
- `offset` (int): Skip events (default: 0)

**Response (200):**
```json
{
  "events": [
    {
      "id": "evt_123",
      "trace_id": "trace_456",
      "operation_type": "chat",
      "duration_ms": 1000,
      "status": "success",
      "tokens": {
        "total": 25
      },
      "cost": 0.0005,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 150,
  "limit": 100,
  "offset": 0
}
```

---

### Analytics

#### GET /api/projects/:id/analytics/summary
Get analytics summary for a project.

**Query Parameters:**
- `period` (string): Time period (hour, day, week, month)
- `start_date` (ISO 8601): Start date
- `end_date` (ISO 8601): End date

**Response (200):**
```json
{
  "summary": {
    "total_requests": 10000,
    "total_tokens": 500000,
    "total_cost": 25.50,
    "average_latency_ms": 350,
    "error_rate": 0.02,
    "success_rate": 0.98,
    "unique_users": 500,
    "top_models": [
      {
        "model": "gpt-3.5-turbo",
        "count": 8000,
        "percentage": 0.8
      }
    ],
    "top_operations": [
      {
        "operation": "chat",
        "count": 7000,
        "percentage": 0.7
      }
    ]
  }
}
```

#### GET /api/projects/:id/analytics/costs
Get cost breakdown.

**Response (200):**
```json
{
  "costs": {
    "total": 25.50,
    "by_provider": [
      {
        "provider": "openai",
        "cost": 20.00,
        "percentage": 0.78
      }
    ],
    "by_model": [
      {
        "model": "gpt-3.5-turbo",
        "cost": 15.00,
        "percentage": 0.59
      }
    ],
    "by_day": [
      {
        "date": "2024-01-01",
        "cost": 5.50
      }
    ]
  }
}
```

#### GET /api/projects/:id/analytics/latency
Get latency distribution.

**Response (200):**
```json
{
  "latency": {
    "p50": 250,
    "p90": 500,
    "p95": 750,
    "p99": 1500,
    "average": 350,
    "min": 50,
    "max": 5000,
    "distribution": [
      {
        "bucket": "0-100ms",
        "count": 1000,
        "percentage": 0.1
      }
    ]
  }
}
```

#### GET /api/projects/:id/analytics/errors
Get error analysis.

**Response (200):**
```json
{
  "errors": {
    "total": 200,
    "rate": 0.02,
    "by_type": [
      {
        "type": "rate_limit",
        "count": 100,
        "percentage": 0.5
      }
    ],
    "by_model": [
      {
        "model": "gpt-4",
        "count": 150,
        "error_rate": 0.03
      }
    ],
    "recent": [
      {
        "id": "evt_error_123",
        "error": "Rate limit exceeded",
        "timestamp": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

---

### Evaluations

#### POST /api/projects/:id/evaluations
Create a new evaluation.

**Request Body:**
```json
{
  "name": "Customer Service Prompt Evaluation",
  "description": "Evaluate the quality of customer service responses",
  "prompt": "You are a helpful customer service assistant..."
}
```

**Response (201):**
```json
{
  "evaluation": {
    "id": 1,
    "project_id": 1,
    "name": "Customer Service Prompt Evaluation",
    "status": "pending",
    "progress": 0,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### GET /api/projects/:id/evaluations
List evaluations for a project.

**Query Parameters:**
- `status` (string): Filter by status (pending, running, completed, failed)
- `limit` (int): Number of results
- `offset` (int): Skip results

**Response (200):**
```json
{
  "evaluations": [
    {
      "id": 1,
      "name": "Customer Service Prompt Evaluation",
      "status": "completed",
      "progress": 100,
      "overall_score": 0.85,
      "created_at": "2024-01-01T00:00:00Z",
      "completed_at": "2024-01-01T00:05:00Z"
    }
  ]
}
```

#### POST /api/evaluations/:id/run
Run an evaluation.

**Request Body:**
```json
{
  "async": true
}
```

**Response (200):**
```json
{
  "message": "Evaluation started",
  "evaluation_id": 1
}
```

#### GET /api/evaluations/:id
Get evaluation details.

**Response (200):**
```json
{
  "evaluation": {
    "id": 1,
    "name": "Customer Service Prompt Evaluation",
    "status": "completed",
    "progress": 100,
    "prompt_analysis": {
      "task_type": "generation",
      "confidence": 0.95,
      "input_schema": {},
      "output_schema": {}
    },
    "test_cases": [
      {
        "id": "test_1",
        "name": "Greeting Test",
        "status": "passed",
        "score": 0.9
      }
    ],
    "metrics": {
      "overall_score": 0.85,
      "pass_rate": 0.9,
      "test_cases_passed": 9,
      "test_cases_total": 10
    }
  }
}
```

#### GET /api/evaluations/:id/metrics
Get evaluation metrics.

**Response (200):**
```json
{
  "metrics": {
    "overall_score": 0.85,
    "pass_rate": 0.9,
    "test_cases_passed": 9,
    "test_cases_total": 10,
    "classification_metrics": {
      "accuracy": 0.92,
      "precision": 0.89,
      "recall": 0.91,
      "f1_score": 0.90
    },
    "generation_metrics": {
      "bleu_score": 0.75,
      "rouge_score": 0.80,
      "semantic_similarity": 0.88
    }
  }
}
```

#### GET /api/evaluations/:id/suggestions
Get optimization suggestions.

**Response (200):**
```json
{
  "suggestions": [
    {
      "id": "sug_1",
      "type": "prompt_improvement",
      "title": "Add Few-Shot Examples",
      "description": "Including examples can improve response quality",
      "expected_impact": 0.15,
      "confidence": 0.85,
      "priority": "high",
      "old_prompt": "You are a helpful assistant.",
      "new_prompt": "You are a helpful assistant.\n\nExamples:\nUser: Hello\nAssistant: Hi! How can I help you today?"
    }
  ]
}
```

#### DELETE /api/evaluations/:id
Delete an evaluation.

**Response (200):**
```json
{
  "message": "Evaluation deleted"
}
```

---

### WebSocket

#### WS /ws
Connect to real-time metrics stream.

**Connection:**
```javascript
const ws = new WebSocket('ws://localhost:8088/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Metrics update:', data);
};
```

**Message Format:**
```json
{
  "type": "metrics_update",
  "data": {
    "requests_per_second": 10.5,
    "average_latency_ms": 250,
    "active_traces": 5,
    "error_rate": 0.01,
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

---

## Error Responses

All errors follow this format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

### Common Error Codes

- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource already exists
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

---

## Rate Limiting

- Default: 1000 requests per hour per API key
- Burst: 100 requests per minute
- Headers returned:
  - `X-RateLimit-Limit`: Request limit
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Reset timestamp

---

## Pagination

List endpoints support pagination:

```
GET /api/projects?page=2&limit=50
```

Response includes:
```json
{
  "data": [...],
  "total": 500,
  "page": 2,
  "limit": 50,
  "has_next": true,
  "has_prev": true
}
```

---

## Webhooks

Configure webhooks in project settings to receive real-time notifications:

```json
{
  "url": "https://your-app.com/webhook",
  "events": ["evaluation.completed", "error.threshold_exceeded"],
  "secret": "webhook_secret_key"
}
```

### Webhook Events

- `evaluation.completed`: Evaluation finished
- `evaluation.failed`: Evaluation failed
- `error.threshold_exceeded`: Error rate exceeded threshold
- `cost.threshold_exceeded`: Cost exceeded threshold
- `latency.threshold_exceeded`: Latency exceeded threshold

### Webhook Payload

```json
{
  "event": "evaluation.completed",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "evaluation_id": 1,
    "project_id": 1,
    "status": "completed",
    "overall_score": 0.85
  }
}
```

### Webhook Security

Verify webhook signatures using HMAC-SHA256:

```python
import hmac
import hashlib

def verify_webhook(payload, signature, secret):
    expected = hmac.new(
        secret.encode(),
        payload.encode(),
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(expected, signature)
```