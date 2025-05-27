"""
Database package providing an abstract interface for video storage.
"""

from .database import Database
from .marqo import MarqoDatabase

__all__ = ['Database', 'MarqoDatabase']
