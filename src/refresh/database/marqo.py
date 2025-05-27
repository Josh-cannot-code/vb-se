"""
Marqo implementation of the video database interface.
"""

from typing import List
from .database import Database
import marqo
from models.video import Video
import json
from dataclasses import asdict

class MarqoDatabase(Database):
    def __init__(self, host: str):
        """Initialize Marqo database connection.
        
        Args:
            host (str): Marqo server host URL
        """
        self.client = marqo.Client(host)
        self.videoIndexSettings = {
            "type": "structured",
            "textPreprocessing": {
                "splitLength": 5,
                "splitOverlap": 1,
                "splitMethod": "sentence"
            },
            "allFields": [
                {"name": "title", "type": "text"},
                {"name": "description", "type": "text"},
                {"name": "transcript", "type": "text"},
                {"name": "url", "type": "text"},
                {"name": "channel_name", "type": "text"},
                {"name": "channel_id", "type": "text"},
                {"name": "thumbnail", "type": "text"},
                {"name": "upload_date", "type": "text"}
            ],
            "tensorFields": [
                "title",
                "description",
                "transcript"
            ]
        }
        self.channelIndexSettings = {
            "type": "structured",
            "allFields": [
                {"name": "channel_name", "type": "text"},
                {"name": "channel_id", "type": "text"}
            ]
        }
        self.pageSize = 1000
    
    #def getChannelIds(self) -> List[str]:
    #    """Get all channel IDs stored in the database.
    #    
    #    Returns:
    #        List[str]: List of unique channel IDs
    #    """
    #    try:
    #        self.client.get_index("vb-se-channels")
    #    except marqo.errors.MarqoWebError as e:
    #        if e.status_code == 404:
    #            print("Index not found, creating it now. Channels will need to be added manually.")
    #            self.client.create_index("vb-se-channels", settings_dict=self.channelIndexSettings)
    #            return []

    #    result = self.client.index("vb-se-channels").search("*", search_method="LEXICAL")
    #    hits = result["hits"]
    #    channelIds = [hit["_id"] for hit in hits]
    #    return channelIds
    
    def getVideoIdsForChannel(self, channel_id: str) -> List[str]:
        """Get all video IDs for a specific channel.
        
        Args:
            channel_id (str): The channel ID to get videos for
            
        Returns:
            List[str]: List of video IDs belonging to the channel
        """
        try:
            self.client.get_index("vb-se-videos")
        except marqo.errors.MarqoWebError as e:
            if e.status_code == 404:
                print("Index not found, will be created on video insertion")
                return []

        videoIds = []
        while True:
            result = self.client.index("vb-se-videos").search("*", offset=len(videoIds), limit=self.pageSize)
            hits = result["hits"]
            videoIds += [hit["_id"] for hit in hits if hit["channel_id"] == channel_id]
            if len(hits) < self.pageSize:
                break

        return videoIds
    
    def putVideo(self, video: Video) -> bool:
        """Store a video object in the database.
        
        Args:
            video (Video): Video object containing required fields
                
        Returns:
            bool: True if successful, False otherwise
        """
        try:
            self.client.get_index("vb-se-videos")
        except marqo.errors.MarqoWebError as e:
            if e.status_code == 404:
                print("Index not found, creating index")
                self.client.create_index("vb-se-videos", settings_dict=self.videoIndexSettings)

        try:
            self.client.index("vb-se-videos").add_documents([asdict(video)])
            return True
        except Exception as e:
            print(f"Failed to add video to database: {str(e)}")
            return False
