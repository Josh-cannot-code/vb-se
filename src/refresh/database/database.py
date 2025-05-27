"""
Abstract base class for video database implementations.
"""

from abc import ABC, abstractmethod
from typing import List
from models.video import Video

class Database(ABC):
    #@abstractmethod
    #def getChannelIds(self) -> List[str]:
    #    """Get all channel IDs stored in the database.

    #    Returns:
    #        List[str]: List of unique channel IDs
    #    """
    #    pass

    @abstractmethod
    def getVideoIdsForChannel(self, channel_id: str) -> List[str]:
        """Get all video IDs for a specific channel.

        Args:
            channel_id (str): The channel ID to get videos for

        Returns:
            List[str]: List of video IDs belonging to the channel
        """
        pass

    @abstractmethod
    def putVideo(self, video: Video) -> bool:
        """Store a video object in the database.

        Args:
            video (Video): Video object containing:
                - id: str
                - video_id: str
                - title: str
                - description: str
                - upload_date: str
                - channel_id: str
                - channel_name: str
                - thumbnail: str
                - url: str
                - transcript: str

        Returns:
            bool: True if successful, False otherwise
        """
        pass
    

