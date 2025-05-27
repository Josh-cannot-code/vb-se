import os
import dotenv
from database import MarqoDatabase
import json
from models.video import Video

def getYoutubeVideoIdsForChannel(channelId: str, ytApiKey: str) -> list[str]:
    # Get all video IDs from YouTube API
    from googleapiclient.discovery import build
    youtube = build('youtube', 'v3', developerKey=ytApiKey)
    
    # Get uploads playlist ID
    channels_response = youtube.channels().list(
        part="contentDetails",
        id=channelId
    ).execute()
        
    if not channels_response['items']:
        print(f"Channel {channelId} not found")
        return []
        
    uploads_playlist_id = channels_response['items'][0]['contentDetails']['relatedPlaylists']['uploads']
        
    # Get all videos from the uploads playlist
    allVideoIds = []
    next_page_token = None
        
    while True:
        playlist_response = youtube.playlistItems().list(
            part="contentDetails",
            playlistId=uploads_playlist_id,
            maxResults=50,
            pageToken=next_page_token
        ).execute()
        
        for item in playlist_response['items']:
            allVideoIds.append(item['contentDetails']['videoId'])
        
        next_page_token = playlist_response.get('nextPageToken')
        if not next_page_token:
            break

    return allVideoIds


def main():
    # Load environment variables
    dotenv.load_dotenv("../.env")

    # Get environment variables
    ytApiKey = os.getenv("YOUTUBE_API_KEY")
    marqoHost = os.getenv("MARQO_HOST")
    channelsJsonList = os.getenv("CHANNELS")
    
    if marqoHost is None:
        raise ValueError("MARQO_HOST must be set")
    if ytApiKey is None:
        raise ValueError("YOUTUBE_API_KEY must be set")
    if channelsJsonList is None:
        raise ValueError("CHANNELS must be set")
    
    print(f"Using Marqo host: {marqoHost}")

    # Initialize database
    db = MarqoDatabase(marqoHost)
    
    # Parse channel IDs
    channelIds: List[str] = json.loads(channelsJsonList)
    
    # For each channel, get all video IDs
    for channelId in channelIds:
        print(f"Processing channel {channelId}")
        videoIds = db.getVideoIdsForChannel(channelId)
        print(f"Found {len(videoIds)} in database")

        # Get all video IDs from YouTube API
        youtubeVideoIds = getYoutubeVideoIdsForChannel(channelId, ytApiKey)
        print(f"Found {len(youtubeVideoIds)} total videos for channel")
    
        # Find videos to process (videos in YouTube but not in database)
        videosToProcess = [vid for vid in youtubeVideoIds if vid not in videoIds]
        print(f"Need to process {len(videosToProcess)} new videos")
    
        for videoId in videosToProcess:
            print(f"Processing video: {videoId}")

            video = Video(videoId)
            
            if db.putVideo(video):
                print(f"Successfully indexed video {videoId}")
            else:
                print(f"Failed to index video {videoId}")

    print("Video refresh process completed")

if __name__ == "__main__":
    main()