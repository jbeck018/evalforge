"""
EvalForge Python SDK

Zero-overhead LLM observability and evaluation platform.
"""

__version__ = "0.1.0"

from .client import EvalForge
from .wrappers import wrap_openai, wrap_anthropic
from .decorators import trace
from .models import TraceEvent, TokenUsage
from .config import configure

__all__ = [
    "EvalForge",
    "wrap_openai", 
    "wrap_anthropic",
    "trace",
    "TraceEvent",
    "TokenUsage",
    "configure",
]