"""
EvalForge Python SDK Client
"""

import asyncio
import json
import threading
import time
import uuid
from typing import Dict, List, Optional, Any, Union
from datetime import datetime, timezone
import queue
import requests
from asyncio_throttle import Throttler

from .config import Config
from .models import TraceEvent, TokenUsage


class EvalForge:
    """
    EvalForge client for event ingestion and analytics.
    
    This client provides zero-overhead observability by batching events
    and sending them asynchronously to minimize impact on application performance.
    """

    def __init__(
        self,
        api_key: str,
        project_id: int,
        base_url: str = "https://api.evalforge.dev",
        batch_size: int = 100,
        flush_interval: float = 5.0,
        max_retries: int = 3,
        timeout: float = 30.0,
        debug: bool = False,
    ):
        """
        Initialize EvalForge client.

        Args:
            api_key: Your EvalForge API key
            project_id: Project ID for event attribution
            base_url: API base URL (default: https://api.evalforge.dev)
            batch_size: Number of events to batch before sending (default: 100)
            flush_interval: Time in seconds between automatic flushes (default: 5.0)
            max_retries: Maximum number of retry attempts (default: 3)
            timeout: Request timeout in seconds (default: 30.0)
            debug: Enable debug logging (default: False)
        """
        self.api_key = api_key
        self.project_id = project_id
        self.base_url = base_url.rstrip("/")
        self.batch_size = batch_size
        self.flush_interval = flush_interval
        self.max_retries = max_retries
        self.timeout = timeout
        self.debug = True  # Force debug mode for testing

        # Event buffer and threading
        self._event_queue: queue.Queue = queue.Queue()
        self._background_thread: Optional[threading.Thread] = None
        self._stop_event = threading.Event()
        self._flush_event = threading.Event()
        
        # Rate limiting (1000 requests per minute)
        self._throttler = Throttler(rate_limit=1000, period=60)
        
        # Start background thread
        self._start_background_thread()

    def _start_background_thread(self):
        """Start the background thread for event processing."""
        if self._background_thread is None or not self._background_thread.is_alive():
            self._background_thread = threading.Thread(
                target=self._background_worker,
                daemon=True,
                name="evalforge-worker"
            )
            self._background_thread.start()

    def _background_worker(self):
        """Background worker that processes events in batches."""
        batch = []
        last_flush = time.time()

        while not self._stop_event.is_set():
            try:
                # Try to get an event with timeout
                try:
                    event = self._event_queue.get(timeout=1.0)
                    batch.append(event)
                    self._event_queue.task_done()
                except queue.Empty:
                    pass

                current_time = time.time()
                should_flush = (
                    len(batch) >= self.batch_size or
                    (batch and current_time - last_flush >= self.flush_interval) or
                    self._flush_event.is_set()
                )

                if should_flush and batch:
                    self._send_batch(batch)
                    batch = []
                    last_flush = current_time
                    if self._flush_event.is_set():
                        self._flush_event.clear()

            except Exception as e:
                if self.debug:
                    print(f"EvalForge background worker error: {e}")

        # Final flush when stopping
        if batch:
            self._send_batch(batch)

    def _send_batch(self, events: List[TraceEvent]):
        """Send a batch of events to the API."""
        if not events:
            return

        url = f"{self.base_url}/sdk/v1/projects/{self.project_id}/events/batch"
        headers = {
            "X-API-Key": self.api_key,
            "Content-Type": "application/json",
        }

        payload = {
            "events": [event.model_dump(mode='json') for event in events]
        }

        for attempt in range(self.max_retries + 1):
            try:
                if self.debug:
                    print(f"EvalForge: Sending request to {url}")
                    print(f"EvalForge: Headers: {headers}")
                response = requests.post(
                    url,
                    json=payload,
                    headers=headers,
                    timeout=self.timeout
                )
                
                if response.status_code == 201:
                    if self.debug:
                        print(f"EvalForge: Successfully sent {len(events)} events")
                    return
                elif response.status_code == 429:  # Rate limited
                    if attempt < self.max_retries:
                        wait_time = 2 ** attempt
                        time.sleep(wait_time)
                        continue
                else:
                    if self.debug:
                        print(f"EvalForge: API error {response.status_code}: {response.text}")
                    return
                    
            except requests.exceptions.RequestException as e:
                if attempt < self.max_retries:
                    wait_time = 2 ** attempt
                    time.sleep(wait_time)
                    continue
                else:
                    if self.debug:
                        print(f"EvalForge: Failed to send events after {self.max_retries} retries: {e}")

    def trace(
        self,
        operation_type: str,
        input_data: Optional[Dict[str, Any]] = None,
        output_data: Optional[Dict[str, Any]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        trace_id: Optional[str] = None,
        span_id: Optional[str] = None,
        parent_span_id: Optional[str] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        status: str = "success",
        tokens: Optional[TokenUsage] = None,
        cost: Optional[float] = None,
        provider: Optional[str] = None,
        model: Optional[str] = None,
    ) -> str:
        """
        Record a trace event.

        Args:
            operation_type: Type of operation (e.g., 'llm_call', 'completion', 'embedding')
            input_data: Input data for the operation
            output_data: Output data from the operation
            metadata: Additional metadata
            trace_id: Trace ID (generated if not provided)
            span_id: Span ID (generated if not provided)
            parent_span_id: Parent span ID for nested operations
            start_time: Operation start time (current time if not provided)
            end_time: Operation end time (current time if not provided)
            status: Operation status ('success', 'error', 'timeout')
            tokens: Token usage information
            cost: Cost in USD
            provider: LLM provider (e.g., 'openai', 'anthropic')
            model: Model name

        Returns:
            The trace ID for this event
        """
        now = datetime.now(timezone.utc)
        trace_id = trace_id or str(uuid.uuid4())
        span_id = span_id or str(uuid.uuid4())
        start_time = start_time or now
        end_time = end_time or now

        duration_ms = int((end_time - start_time).total_seconds() * 1000)

        event = TraceEvent(
            id=str(uuid.uuid4()),
            project_id=self.project_id,
            trace_id=trace_id,
            span_id=span_id,
            parent_span_id=parent_span_id,
            operation_type=operation_type,
            start_time=start_time,
            end_time=end_time,
            duration_ms=duration_ms,
            status=status,
            input=input_data or {},
            output=output_data or {},
            metadata=metadata or {},
            tokens=tokens or TokenUsage(),
            cost=cost or 0.0,
            provider=provider or "",
            model=model or "",
        )

        # Add to queue for background processing
        try:
            self._event_queue.put_nowait(event)
        except queue.Full:
            if self.debug:
                print("EvalForge: Event queue is full, dropping event")

        return trace_id

    def flush(self, timeout: float = 10.0) -> bool:
        """
        Flush all pending events.

        Args:
            timeout: Maximum time to wait for flush completion

        Returns:
            True if flush completed successfully, False if timeout
        """
        self._flush_event.set()
        
        # Wait for queue to be empty
        start_time = time.time()
        while not self._event_queue.empty() and time.time() - start_time < timeout:
            time.sleep(0.1)
        
        return self._event_queue.empty()

    def close(self):
        """Close the client and stop background processing."""
        self._stop_event.set()
        self.flush(timeout=5.0)
        
        if self._background_thread and self._background_thread.is_alive():
            self._background_thread.join(timeout=5.0)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()


# Global client instance
_global_client: Optional[EvalForge] = None


def configure(
    api_key: str,
    project_id: int,
    **kwargs
) -> EvalForge:
    """
    Configure the global EvalForge client.

    Args:
        api_key: Your EvalForge API key
        project_id: Project ID for event attribution
        **kwargs: Additional client configuration options

    Returns:
        The configured EvalForge client
    """
    global _global_client
    _global_client = EvalForge(api_key=api_key, project_id=project_id, **kwargs)
    return _global_client


def get_client() -> Optional[EvalForge]:
    """Get the global EvalForge client."""
    return _global_client


def trace(*args, **kwargs) -> str:
    """
    Record a trace event using the global client.
    
    This is a convenience function that uses the globally configured client.
    """
    client = get_client()
    if client is None:
        raise ValueError("EvalForge client not configured. Call evalforge.configure() first.")
    
    return client.trace(*args, **kwargs)


def flush(timeout: float = 10.0) -> bool:
    """
    Flush all pending events using the global client.
    
    Args:
        timeout: Maximum time to wait for flush completion
        
    Returns:
        True if flush completed successfully, False if timeout
    """
    client = get_client()
    if client is None:
        return True
    
    return client.flush(timeout=timeout)