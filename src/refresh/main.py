import os
import dotenv
import yt_dlp
from youtube_transcript_api import YouTubeTranscriptApi
import requests
from requests.auth import HTTPBasicAuth

def get_channel_ids(opensearchHost: str, opensearchUsername: str, opensearchPassword: str) -> list[str]:
    # Make a raw HTTP request to OpenSearch
    url = f"{opensearchHost}/vb-se-channels/_search"
    response = requests.get(url, auth=HTTPBasicAuth(opensearchUsername, opensearchPassword), verify=False)
    
    # Check if the request was successful
    if response.status_code == 200:
        data = response.json()
        channelIds = [hit["_source"]["channel_id"] for hit in data["hits"]["hits"]]
        print("Got channel IDs: ", channelIds)
        return channelIds
    else:
        raise Exception(f"Failed to retrieve channel IDs: {response.status_code} {response.text}")

def get_stored_video_ids_for_channel(channelId: str, opensearchHost: str, opensearchUsername: str, opensearchPassword: str) -> list[str]:
    # Make a raw HTTP request to OpenSearch
    url = f"{opensearchHost}/vb-se-videos/_search"
    query = {
        "size": 10000,
        "query": {
            "term": {
                "channel_id": channelId
            }
        },
        "_source": ["video_id"]
    }
    response = requests.get(
        url, 
        json=query,
        auth=HTTPBasicAuth(opensearchUsername, opensearchPassword), 
        verify=False
    )
    
    # Check if the request was successful
    if response.status_code == 200:
        data = response.json()
        videoIds = [hit["_source"]["video_id"] for hit in data["hits"]["hits"]]
        return videoIds
    else:
        raise Exception(f"Failed to retrieve video IDs for channel {channelId}: {response.status_code} {response.text}")

def get_youtube_video_ids_for_channel(channelId: str, ytApiKey: str) -> list[str]:
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

def get_video_metadata(videoId: str) -> dict:
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

def get_video_transcript(videoId: str) -> str:
    transcript, unretrievable_transcripts = YouTubeTranscriptApi.get_transcripts([videoId])

    if videoId in unretrievable_transcripts:
        return "NO_TRANSCRIPT"
    else:
        s = ""
        for t in transcript[videoId]:
            s += " " + t["text"]
        return s

def main():
    # Load environment variables
    dotenv.load_dotenv("../.env")

    # Get environment variables
    ytApiKey = os.getenv("YOUTUBE_API_KEY")

    opensearchHost = os.getenv("OPENSEARCH_HOST")
    opensearchUsername = os.getenv("OPENSEARCH_USERNAME")
    opensearchPassword = os.getenv("OPENSEARCH_PASSWORD")

    if opensearchHost is None or opensearchUsername is None or opensearchPassword is None:
        raise ValueError("OPENSEARCH_HOST, OPENSEARCH_USERNAME, and OPENSEARCH_PASSWORD must be set")

    if ytApiKey is None:
        raise ValueError("YOUTUBE_API_KEY not set")

    print(opensearchHost)

    # Get channel IDs
    channelIds = get_channel_ids(opensearchHost, opensearchUsername, opensearchPassword)

    for channelId in channelIds:
        print("Getting videos for channel: ", channelId)
        
        # Get stored video IDs for this channel
        videoIds = get_stored_video_ids_for_channel(channelId, opensearchHost, opensearchUsername, opensearchPassword)
        print(f"Found {len(videoIds)} in database")

        # Get all video IDs from YouTube API
        youtubeVideoIds = get_youtube_video_ids_for_channel(channelId, ytApiKey)
        print(f"Found {len(youtubeVideoIds)} total videos for channel")
    
        # Find videos to process (videos in YouTube but not in database)
        videosToProcess = [vid for vid in youtubeVideoIds if vid not in videoIds]
        print(f"Need to process {len(videosToProcess)} new videos")
    
        # Get details and transcripts for each video
        for videoId in videosToProcess:
            print(f"Processing video: {videoId}")
            
            # Get video details
            video = get_video_metadata(videoId)
            if video is None:
                continue
            
            # Get transcript
            transcript = get_video_transcript(videoId)
            
            if transcript == "NO_TRANSCRIPT":
                print(f"Failed to get transcript for {videoId}")
            
            # Create video object
            video = {
                "id": videoId,
                "video_id": videoId,
                "title": video.get('title', ''),
                "description": video.get('description', ''),
                "upload_date": video.get('upload_date', ''),
                "channel_id": channelId,
                "channel_name": video.get('channel', ''),
                "thumbnail": video.get('thumbnail', ''),
                "url": f"https://www.youtube.com/watch?v={videoId}",
                "transcript": transcript
            }
            
            # Save to OpenSearch
            index_url = f"{opensearchHost}/vb-se-videos/_doc/{videoId}"
            index_response = requests.put(
                index_url, 
                json=video, 
                auth=HTTPBasicAuth(opensearchUsername, opensearchPassword), 
                verify=False
            )
            
            if index_response.status_code in [200, 201]:
                print(f"Successfully indexed video {videoId}")
            else:
                print(f"Failed to index video {videoId}: {index_response.status_code} {index_response.text}")

    print("Video refresh process completed")

if __name__ == "__main__":
    main()