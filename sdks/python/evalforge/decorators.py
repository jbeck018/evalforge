"""
EvalForge Decorators

Decorators for automatically tracing functions and methods.
"""

import functools
import inspect
import uuid
from datetime import datetime, timezone
from typing import Any, Callable, Dict, Optional

from .client import get_client
from .models import TokenUsage


def trace(
    operation_type: Optional[str] = None,
    capture_args: bool = True,
    capture_result: bool = True,
    capture_exceptions: bool = True,
    include_private_args: bool = False,
    provider: Optional[str] = None,
    model: Optional[str] = None,
):
    """
    Decorator to automatically trace function calls.
    
    Args:
        operation_type: Type of operation (defaults to function name)
        capture_args: Whether to capture function arguments (default: True)
        capture_result: Whether to capture function result (default: True)
        capture_exceptions: Whether to capture exceptions (default: True)
        include_private_args: Whether to include arguments starting with _ (default: False)
        provider: LLM provider name
        model: Model name
    
    Usage:
        @evalforge.trace(operation_type="custom_llm_call")
        def my_llm_function(prompt: str) -> str:
            # Your LLM call logic here
            return result
            
        @evalforge.trace()  # Uses function name as operation_type
        def process_data(data: dict) -> dict:
            # Your processing logic here
            return processed_data
    """
    
    def decorator(func: Callable) -> Callable:
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            client = get_client()
            if not client:
                return func(*args, **kwargs)
            
            # Generate trace identifiers
            trace_id = str(uuid.uuid4())
            span_id = str(uuid.uuid4())
            start_time = datetime.now(timezone.utc)
            
            # Determine operation type
            op_type = operation_type or func.__name__
            
            # Capture function arguments
            input_data = {}
            if capture_args:
                # Get function signature
                sig = inspect.signature(func)
                bound_args = sig.bind(*args, **kwargs)
                bound_args.apply_defaults()
                
                for param_name, param_value in bound_args.arguments.items():
                    # Skip private arguments unless explicitly requested
                    if not include_private_args and param_name.startswith('_'):
                        continue
                    
                    # Try to serialize the argument, fallback to string representation
                    try:
                        if isinstance(param_value, (str, int, float, bool, list, dict, type(None))):
                            input_data[param_name] = param_value
                        else:
                            input_data[param_name] = str(param_value)
                    except Exception:
                        input_data[param_name] = f"<{type(param_value).__name__}>"
            
            try:
                # Execute the function
                result = func(*args, **kwargs)
                end_time = datetime.now(timezone.utc)
                
                # Capture result
                output_data = {}
                if capture_result:
                    try:
                        if isinstance(result, (str, int, float, bool, list, dict, type(None))):
                            output_data["result"] = result
                        else:
                            output_data["result"] = str(result)
                    except Exception:
                        output_data["result"] = f"<{type(result).__name__}>"
                
                # Create trace event
                client.trace(
                    operation_type=op_type,
                    input_data=input_data,
                    output_data=output_data,
                    metadata={
                        "function": func.__name__,
                        "module": func.__module__,
                        "qualified_name": func.__qualname__,
                    },
                    trace_id=trace_id,
                    span_id=span_id,
                    start_time=start_time,
                    end_time=end_time,
                    status="success",
                    provider=provider or "",
                    model=model or "",
                )
                
                return result
                
            except Exception as e:
                end_time = datetime.now(timezone.utc)
                
                # Capture exception if enabled
                error_data = {}
                if capture_exceptions:
                    error_data = {
                        "error": str(e),
                        "error_type": type(e).__name__,
                    }
                
                # Create trace event for error
                client.trace(
                    operation_type=op_type,
                    input_data=input_data,
                    output_data=error_data,
                    metadata={
                        "function": func.__name__,
                        "module": func.__module__,
                        "qualified_name": func.__qualname__,
                    },
                    trace_id=trace_id,
                    span_id=span_id,
                    start_time=start_time,
                    end_time=end_time,
                    status="error",
                    provider=provider or "",
                    model=model or "",
                )
                
                raise
        
        return wrapper
    
    return decorator


class TraceContext:
    """
    Context manager for creating nested traces.
    
    Usage:
        with TraceContext("data_processing") as ctx:
            # Operations within this context will be traced
            # as child spans of the main trace
            
            with ctx.child("step_1"):
                # Nested operation
                pass
                
            with ctx.child("step_2"):
                # Another nested operation
                pass
    """
    
    def __init__(
        self,
        operation_type: str,
        input_data: Optional[Dict[str, Any]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        provider: Optional[str] = None,
        model: Optional[str] = None,
        parent_span_id: Optional[str] = None,
    ):
        self.operation_type = operation_type
        self.input_data = input_data or {}
        self.metadata = metadata or {}
        self.provider = provider or ""
        self.model = model or ""
        self.parent_span_id = parent_span_id
        
        self.trace_id = str(uuid.uuid4())
        self.span_id = str(uuid.uuid4())
        self.start_time: Optional[datetime] = None
        self.end_time: Optional[datetime] = None
        
    def __enter__(self):
        self.start_time = datetime.now(timezone.utc)
        return self
        
    def __exit__(self, exc_type, exc_val, exc_tb):
        client = get_client()
        if not client:
            return
            
        self.end_time = datetime.now(timezone.utc)
        
        status = "error" if exc_type is not None else "success"
        output_data = {}
        
        if exc_type is not None:
            output_data = {
                "error": str(exc_val),
                "error_type": exc_type.__name__,
            }
        
        client.trace(
            operation_type=self.operation_type,
            input_data=self.input_data,
            output_data=output_data,
            metadata=self.metadata,
            trace_id=self.trace_id,
            span_id=self.span_id,
            parent_span_id=self.parent_span_id,
            start_time=self.start_time,
            end_time=self.end_time,
            status=status,
            provider=self.provider,
            model=self.model,
        )
    
    def child(
        self,
        operation_type: str,
        input_data: Optional[Dict[str, Any]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        provider: Optional[str] = None,
        model: Optional[str] = None,
    ) -> 'TraceContext':
        """Create a child trace context."""
        return TraceContext(
            operation_type=operation_type,
            input_data=input_data,
            metadata=metadata,
            provider=provider or self.provider,
            model=model or self.model,
            parent_span_id=self.span_id,
        )
    
    def set_output(self, output_data: Dict[str, Any]):
        """Set output data for this trace."""
        self.output_data = output_data
    
    def add_metadata(self, key: str, value: Any):
        """Add metadata to this trace."""
        self.metadata[key] = value