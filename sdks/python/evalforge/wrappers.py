"""
EvalForge LLM Provider Wrappers

These wrappers provide seamless integration with popular LLM providers
while automatically capturing traces and metrics.
"""

import functools
import inspect
import time
import uuid
from datetime import datetime, timezone
from typing import Any, Dict, Optional, Union, Callable

from .client import get_client
from .models import TokenUsage


def _calculate_openai_cost(model: str, tokens: TokenUsage) -> float:
    """Calculate cost for OpenAI models based on current pricing."""
    
    # OpenAI pricing per 1K tokens (as of late 2023)
    pricing = {
        # GPT-4 models
        "gpt-4": {"prompt": 0.03, "completion": 0.06},
        "gpt-4-32k": {"prompt": 0.06, "completion": 0.12},
        "gpt-4-1106-preview": {"prompt": 0.01, "completion": 0.03},
        "gpt-4-0125-preview": {"prompt": 0.01, "completion": 0.03},
        
        # GPT-3.5 models
        "gpt-3.5-turbo": {"prompt": 0.001, "completion": 0.002},
        "gpt-3.5-turbo-16k": {"prompt": 0.003, "completion": 0.004},
        "gpt-3.5-turbo-1106": {"prompt": 0.001, "completion": 0.002},
        "gpt-3.5-turbo-0125": {"prompt": 0.0005, "completion": 0.0015},
        
        # Embedding models
        "text-embedding-ada-002": {"prompt": 0.0001, "completion": 0},
        "text-embedding-3-small": {"prompt": 0.00002, "completion": 0},
        "text-embedding-3-large": {"prompt": 0.00013, "completion": 0},
    }
    
    if model not in pricing:
        # Default to GPT-3.5 pricing for unknown models
        model_pricing = pricing["gpt-3.5-turbo"]
    else:
        model_pricing = pricing[model]
    
    prompt_cost = (tokens.prompt / 1000) * model_pricing["prompt"]
    completion_cost = (tokens.completion / 1000) * model_pricing["completion"]
    
    return prompt_cost + completion_cost


def _calculate_anthropic_cost(model: str, tokens: TokenUsage) -> float:
    """Calculate cost for Anthropic models based on current pricing."""
    
    # Anthropic pricing per 1K tokens (as of late 2023)
    pricing = {
        "claude-3-opus-20240229": {"prompt": 0.015, "completion": 0.075},
        "claude-3-sonnet-20240229": {"prompt": 0.003, "completion": 0.015},
        "claude-3-haiku-20240307": {"prompt": 0.00025, "completion": 0.00125},
        "claude-2.1": {"prompt": 0.008, "completion": 0.024},
        "claude-2.0": {"prompt": 0.008, "completion": 0.024},
        "claude-instant-1.2": {"prompt": 0.0008, "completion": 0.0024},
    }
    
    if model not in pricing:
        # Default to Claude 3 Sonnet pricing for unknown models
        model_pricing = pricing["claude-3-sonnet-20240229"]
    else:
        model_pricing = pricing[model]
    
    prompt_cost = (tokens.prompt / 1000) * model_pricing["prompt"]
    completion_cost = (tokens.completion / 1000) * model_pricing["completion"]
    
    return prompt_cost + completion_cost


def _extract_openai_tokens(response: Any) -> TokenUsage:
    """Extract token usage from OpenAI response."""
    if hasattr(response, 'usage') and response.usage:
        return TokenUsage(
            prompt=response.usage.prompt_tokens,
            completion=response.usage.completion_tokens,
            total=response.usage.total_tokens
        )
    return TokenUsage()


def _extract_anthropic_tokens(response: Any) -> TokenUsage:
    """Extract token usage from Anthropic response."""
    if hasattr(response, 'usage') and response.usage:
        return TokenUsage(
            prompt=response.usage.input_tokens,
            completion=response.usage.output_tokens,
            total=response.usage.input_tokens + response.usage.output_tokens
        )
    return TokenUsage()


def wrap_openai(client_class):
    """
    Wrap OpenAI client to automatically trace API calls.
    
    Usage:
        import openai
        import evalforge
        
        # Configure EvalForge
        evalforge.configure(api_key="your-key", project_id=1)
        
        # Wrap OpenAI client
        OpenAI = evalforge.wrap_openai(openai.OpenAI)
        client = OpenAI(api_key="your-openai-key")
        
        # Use normally - traces are automatically captured
        response = client.chat.completions.create(
            model="gpt-3.5-turbo",
            messages=[{"role": "user", "content": "Hello!"}]
        )
    """
    
    class WrappedOpenAI(client_class):
        def __init__(self, *args, **kwargs):
            super().__init__(*args, **kwargs)
            
            # Wrap chat completions
            if hasattr(self, 'chat') and hasattr(self.chat, 'completions'):
                original_create = self.chat.completions.create
                self.chat.completions.create = _wrap_openai_chat_completion(original_create)
            
            # Wrap embeddings
            if hasattr(self, 'embeddings'):
                original_create = self.embeddings.create
                self.embeddings.create = _wrap_openai_embedding(original_create)
    
    return WrappedOpenAI


def _wrap_openai_chat_completion(original_func):
    """Wrap OpenAI chat completion method."""
    
    @functools.wraps(original_func)
    def wrapper(*args, **kwargs):
        client = get_client()
        if not client:
            return original_func(*args, **kwargs)
        
        trace_id = str(uuid.uuid4())
        span_id = str(uuid.uuid4())
        start_time = datetime.now(timezone.utc)
        
        # Extract model and messages for tracing
        model = kwargs.get('model', 'unknown')
        messages = kwargs.get('messages', [])
        
        try:
            response = original_func(*args, **kwargs)
            end_time = datetime.now(timezone.utc)
            
            # Extract response data
            tokens = _extract_openai_tokens(response)
            cost = _calculate_openai_cost(model, tokens)
            
            # Get response content
            output_content = ""
            if hasattr(response, 'choices') and response.choices:
                choice = response.choices[0]
                if hasattr(choice, 'message') and hasattr(choice.message, 'content'):
                    output_content = choice.message.content
            
            client.trace(
                operation_type="llm_call",
                input_data={
                    "model": model,
                    "messages": messages,
                    "kwargs": {k: v for k, v in kwargs.items() if k not in ['model', 'messages']}
                },
                output_data={
                    "content": output_content,
                    "finish_reason": getattr(response.choices[0], 'finish_reason', None) if response.choices else None
                },
                metadata={
                    "provider": "openai",
                    "response_id": getattr(response, 'id', None),
                    "system_fingerprint": getattr(response, 'system_fingerprint', None),
                },
                trace_id=trace_id,
                span_id=span_id,
                start_time=start_time,
                end_time=end_time,
                status="success",
                tokens=tokens,
                cost=cost,
                provider="openai",
                model=model,
            )
            
            return response
            
        except Exception as e:
            end_time = datetime.now(timezone.utc)
            
            client.trace(
                operation_type="llm_call",
                input_data={
                    "model": model,
                    "messages": messages,
                    "kwargs": {k: v for k, v in kwargs.items() if k not in ['model', 'messages']}
                },
                output_data={},
                metadata={
                    "provider": "openai",
                    "error": str(e),
                    "error_type": type(e).__name__,
                },
                trace_id=trace_id,
                span_id=span_id,
                start_time=start_time,
                end_time=end_time,
                status="error",
                tokens=TokenUsage(),
                cost=0.0,
                provider="openai",
                model=model,
            )
            
            raise
    
    return wrapper


def _wrap_openai_embedding(original_func):
    """Wrap OpenAI embedding method."""
    
    @functools.wraps(original_func)
    def wrapper(*args, **kwargs):
        client = get_client()
        if not client:
            return original_func(*args, **kwargs)
        
        trace_id = str(uuid.uuid4())
        span_id = str(uuid.uuid4())
        start_time = datetime.now(timezone.utc)
        
        # Extract model and input for tracing
        model = kwargs.get('model', 'unknown')
        input_text = kwargs.get('input', '')
        
        try:
            response = original_func(*args, **kwargs)
            end_time = datetime.now(timezone.utc)
            
            # Extract response data
            tokens = _extract_openai_tokens(response)
            cost = _calculate_openai_cost(model, tokens)
            
            client.trace(
                operation_type="embedding",
                input_data={
                    "model": model,
                    "input": input_text,
                    "kwargs": {k: v for k, v in kwargs.items() if k not in ['model', 'input']}
                },
                output_data={
                    "embedding_count": len(response.data) if hasattr(response, 'data') else 0,
                },
                metadata={
                    "provider": "openai",
                    "response_id": getattr(response, 'id', None),
                },
                trace_id=trace_id,
                span_id=span_id,
                start_time=start_time,
                end_time=end_time,
                status="success",
                tokens=tokens,
                cost=cost,
                provider="openai",
                model=model,
            )
            
            return response
            
        except Exception as e:
            end_time = datetime.now(timezone.utc)
            
            client.trace(
                operation_type="embedding",
                input_data={
                    "model": model,
                    "input": input_text,
                    "kwargs": {k: v for k, v in kwargs.items() if k not in ['model', 'input']}
                },
                output_data={},
                metadata={
                    "provider": "openai",
                    "error": str(e),
                    "error_type": type(e).__name__,
                },
                trace_id=trace_id,
                span_id=span_id,
                start_time=start_time,
                end_time=end_time,
                status="error",
                tokens=TokenUsage(),
                cost=0.0,
                provider="openai",
                model=model,
            )
            
            raise
    
    return wrapper


def wrap_anthropic(client_class):
    """
    Wrap Anthropic client to automatically trace API calls.
    
    Usage:
        import anthropic
        import evalforge
        
        # Configure EvalForge
        evalforge.configure(api_key="your-key", project_id=1)
        
        # Wrap Anthropic client
        Anthropic = evalforge.wrap_anthropic(anthropic.Anthropic)
        client = Anthropic(api_key="your-anthropic-key")
        
        # Use normally - traces are automatically captured
        response = client.messages.create(
            model="claude-3-sonnet-20240229",
            max_tokens=1024,
            messages=[{"role": "user", "content": "Hello!"}]
        )
    """
    
    class WrappedAnthropic(client_class):
        def __init__(self, *args, **kwargs):
            super().__init__(*args, **kwargs)
            
            # Wrap messages
            if hasattr(self, 'messages'):
                original_create = self.messages.create
                self.messages.create = _wrap_anthropic_message(original_create)
    
    return WrappedAnthropic


def _wrap_anthropic_message(original_func):
    """Wrap Anthropic message method."""
    
    @functools.wraps(original_func)
    def wrapper(*args, **kwargs):
        client = get_client()
        if not client:
            return original_func(*args, **kwargs)
        
        trace_id = str(uuid.uuid4())
        span_id = str(uuid.uuid4())
        start_time = datetime.now(timezone.utc)
        
        # Extract model and messages for tracing
        model = kwargs.get('model', 'unknown')
        messages = kwargs.get('messages', [])
        max_tokens = kwargs.get('max_tokens', 0)
        
        try:
            response = original_func(*args, **kwargs)
            end_time = datetime.now(timezone.utc)
            
            # Extract response data
            tokens = _extract_anthropic_tokens(response)
            cost = _calculate_anthropic_cost(model, tokens)
            
            # Get response content
            output_content = ""
            if hasattr(response, 'content') and response.content:
                if isinstance(response.content, list) and response.content:
                    first_content = response.content[0]
                    if hasattr(first_content, 'text'):
                        output_content = first_content.text
                elif hasattr(response.content, 'text'):
                    output_content = response.content.text
            
            client.trace(
                operation_type="llm_call",
                input_data={
                    "model": model,
                    "messages": messages,
                    "max_tokens": max_tokens,
                    "kwargs": {k: v for k, v in kwargs.items() if k not in ['model', 'messages', 'max_tokens']}
                },
                output_data={
                    "content": output_content,
                    "stop_reason": getattr(response, 'stop_reason', None),
                },
                metadata={
                    "provider": "anthropic",
                    "response_id": getattr(response, 'id', None),
                    "response_type": getattr(response, 'type', None),
                },
                trace_id=trace_id,
                span_id=span_id,
                start_time=start_time,
                end_time=end_time,
                status="success",
                tokens=tokens,
                cost=cost,
                provider="anthropic",
                model=model,
            )
            
            return response
            
        except Exception as e:
            end_time = datetime.now(timezone.utc)
            
            client.trace(
                operation_type="llm_call",
                input_data={
                    "model": model,
                    "messages": messages,
                    "max_tokens": max_tokens,
                    "kwargs": {k: v for k, v in kwargs.items() if k not in ['model', 'messages', 'max_tokens']}
                },
                output_data={},
                metadata={
                    "provider": "anthropic",
                    "error": str(e),
                    "error_type": type(e).__name__,
                },
                trace_id=trace_id,
                span_id=span_id,
                start_time=start_time,
                end_time=end_time,
                status="error",
                tokens=TokenUsage(),
                cost=0.0,
                provider="anthropic",
                model=model,
            )
            
            raise
    
    return wrapper