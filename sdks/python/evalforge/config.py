"""
EvalForge Configuration
"""

import os
from typing import Optional


class Config:
    """Configuration class for EvalForge SDK."""
    
    def __init__(self):
        self.api_key: Optional[str] = None
        self.project_id: Optional[int] = None
        self.base_url: str = "https://api.evalforge.dev"
        self.batch_size: int = 100
        self.flush_interval: float = 5.0
        self.max_retries: int = 3
        self.timeout: float = 30.0
        self.debug: bool = False
        
    @classmethod
    def from_env(cls) -> 'Config':
        """Load configuration from environment variables."""
        config = cls()
        
        config.api_key = os.getenv('EVALFORGE_API_KEY')
        
        project_id = os.getenv('EVALFORGE_PROJECT_ID')
        if project_id:
            try:
                config.project_id = int(project_id)
            except ValueError:
                raise ValueError(f"Invalid EVALFORGE_PROJECT_ID: {project_id}")
        
        config.base_url = os.getenv('EVALFORGE_BASE_URL', config.base_url)
        
        batch_size = os.getenv('EVALFORGE_BATCH_SIZE')
        if batch_size:
            try:
                config.batch_size = int(batch_size)
            except ValueError:
                raise ValueError(f"Invalid EVALFORGE_BATCH_SIZE: {batch_size}")
        
        flush_interval = os.getenv('EVALFORGE_FLUSH_INTERVAL')
        if flush_interval:
            try:
                config.flush_interval = float(flush_interval)
            except ValueError:
                raise ValueError(f"Invalid EVALFORGE_FLUSH_INTERVAL: {flush_interval}")
        
        max_retries = os.getenv('EVALFORGE_MAX_RETRIES')
        if max_retries:
            try:
                config.max_retries = int(max_retries)
            except ValueError:
                raise ValueError(f"Invalid EVALFORGE_MAX_RETRIES: {max_retries}")
        
        timeout = os.getenv('EVALFORGE_TIMEOUT')
        if timeout:
            try:
                config.timeout = float(timeout)
            except ValueError:
                raise ValueError(f"Invalid EVALFORGE_TIMEOUT: {timeout}")
        
        debug = os.getenv('EVALFORGE_DEBUG', '').lower()
        config.debug = debug in ('true', '1', 'yes', 'on')
        
        return config


# Global configuration instance
_config = Config()


def configure(
    api_key: Optional[str] = None,
    project_id: Optional[int] = None,
    base_url: Optional[str] = None,
    batch_size: Optional[int] = None,
    flush_interval: Optional[float] = None,
    max_retries: Optional[int] = None,
    timeout: Optional[float] = None,
    debug: Optional[bool] = None,
):
    """
    Configure the global EvalForge settings.
    
    Args:
        api_key: Your EvalForge API key
        project_id: Project ID for event attribution
        base_url: API base URL
        batch_size: Number of events to batch before sending
        flush_interval: Time in seconds between automatic flushes
        max_retries: Maximum number of retry attempts
        timeout: Request timeout in seconds
        debug: Enable debug logging
    """
    global _config
    
    if api_key is not None:
        _config.api_key = api_key
    if project_id is not None:
        _config.project_id = project_id
    if base_url is not None:
        _config.base_url = base_url
    if batch_size is not None:
        _config.batch_size = batch_size
    if flush_interval is not None:
        _config.flush_interval = flush_interval
    if max_retries is not None:
        _config.max_retries = max_retries
    if timeout is not None:
        _config.timeout = timeout
    if debug is not None:
        _config.debug = debug


def get_config() -> Config:
    """Get the global configuration instance."""
    return _config


def load_from_env():
    """Load configuration from environment variables."""
    global _config
    _config = Config.from_env()