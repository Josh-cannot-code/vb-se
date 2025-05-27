"""
Video data model.
"""

from typing import Optional
import yt_dlp
from youtube_transcript_api import YouTubeTranscriptApi
from dataclasses import dataclass
from datetime import datetime

@dataclass(init=False)
class Video:
    """Video data model representing a YouTube video."""
    _id: str
    title: str
    description: str
    upload_date: str
    channel_id: str
    channel_name: str
    thumbnail: str
    url: str
    transcript: Optional[str] = None

    def __init__(self, videoId: str):
        video = self.__getVideoMetadata(videoId)
        if video is None:
            raise ValueError(f"Failed to get metadata for video {videoId}")
        
        transcript = self.__getVideoTranscript(videoId)
        if transcript is None:
            transcript = "NO_TRANSCRIPT"
        
        originalTimeFormat = "%Y%m%d"
        outputFormat = "%Y-%m-%dT%H:%M:%SZ"
        dt = datetime.strptime(video['upload_date'], originalTimeFormat)
        upload_date = dt.strftime(outputFormat)

        self._id = videoId
        self.title = video['title']
        self.description = video['description']
        self.upload_date = upload_date
        self.channel_id = video['channel_id']
        self.channel_name = video['channel_name']
        self.thumbnail = video['thumbnail']
        self.url = video['url']
        self.transcript = transcript
        
    def __getVideoMetadata(self, videoId: str) -> dict:
        options = {
                "quiet": True,
        }

        with yt_dlp.YoutubeDL(options) as ydl:
            try:
                result = ydl.extract_info('https://www.youtube.com/watch?v=' + videoId, download=False)
            except Exception as e:
                print(f"Failed to get metadata for video {videoId}: {str(e)}")
                return None

        video = {
            'video_id': result['id'],
            'title': result['title'],
            'thumbnail': result['thumbnail'],
            'channel_id': result['channel_id'],
            'description': result['description'],
            'upload_date': result['upload_date'],
            'url': result['original_url'],
            'channel_name': result['channel']
        }

        return video

    def __getVideoTranscript(self, videoId: str) -> Optional[str]:
        yttapi = YouTubeTranscriptApi()
        try:
            transcript = yttapi.fetch(videoId)
        except Exception as e:
            print(f"Failed to get transcript for video {videoId}: {str(e)}")
            return None

        s = ""
        for t in transcript:
            s += " " + t.text

        return s