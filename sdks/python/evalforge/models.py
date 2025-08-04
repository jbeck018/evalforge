"""
EvalForge SDK Models
"""

from datetime import datetime
from typing import Dict, Any, Optional
from pydantic import BaseModel, Field


class TokenUsage(BaseModel):
    """Token usage information for LLM calls."""
    
    prompt: int = Field(default=0, description="Number of prompt tokens")
    completion: int = Field(default=0, description="Number of completion tokens")
    total: int = Field(default=0, description="Total number of tokens")

    def __init__(self, **data):
        super().__init__(**data)
        # Auto-calculate total if not provided
        if self.total == 0:
            self.total = self.prompt + self.completion


class TraceEvent(BaseModel):
    """A trace event representing an operation in the system."""
    
    id: str = Field(description="Unique event ID")
    project_id: int = Field(description="Project ID")
    trace_id: str = Field(description="Trace ID for grouping related events")
    span_id: str = Field(description="Unique span ID")
    parent_span_id: Optional[str] = Field(default=None, description="Parent span ID for nested operations")
    operation_type: str = Field(description="Type of operation (e.g., 'llm_call', 'completion')")
    start_time: datetime = Field(description="Operation start time")
    end_time: datetime = Field(description="Operation end time")
    duration_ms: int = Field(description="Operation duration in milliseconds")
    status: str = Field(default="success", description="Operation status")
    input: Dict[str, Any] = Field(default_factory=dict, description="Input data")
    output: Dict[str, Any] = Field(default_factory=dict, description="Output data")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Additional metadata")
    tokens: TokenUsage = Field(default_factory=TokenUsage, description="Token usage information")
    cost: float = Field(default=0.0, description="Cost in USD")
    provider: str = Field(default="", description="LLM provider")
    model: str = Field(default="", description="Model name")

    class Config:
        json_encoders = {
            datetime: lambda v: v.isoformat()
        }