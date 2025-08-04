#!/usr/bin/env python3
"""
EvalForge Python SDK
"""

from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

with open("requirements.txt", "r", encoding="utf-8") as fh:
    requirements = [line.strip() for line in fh if line.strip() and not line.startswith("#")]

setup(
    name="evalforge",
    version="0.1.0",
    author="EvalForge Team",
    author_email="team@evalforge.dev",
    description="EvalForge Python SDK - Zero-overhead LLM observability and evaluation",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/evalforge/evalforge",
    project_urls={
        "Bug Tracker": "https://github.com/evalforge/evalforge/issues",
        "Documentation": "https://docs.evalforge.dev",
    },
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: Scientific/Engineering :: Artificial Intelligence",
    ],
    packages=find_packages(),
    python_requires=">=3.8",
    install_requires=requirements,
    extras_require={
        "dev": [
            "pytest>=7.0.0",
            "pytest-asyncio>=0.21.0",
            "pytest-cov>=4.0.0",
            "black>=23.0.0",
            "isort>=5.12.0",
            "flake8>=6.0.0",
            "mypy>=1.0.0",
            "pre-commit>=3.0.0",
        ],
        "anthropic": ["anthropic>=0.7.0"],
        "openai": ["openai>=1.0.0"],
    },
    entry_points={
        "console_scripts": [
            "evalforge=evalforge.cli:main",
        ],
    },
)