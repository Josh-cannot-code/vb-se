from __future__ import annotations
from time import sleep
from googleapiclient.discovery import build  # pyright: ignore[reportUnknownVariableType]
from youtube_transcript_api import YouTubeTranscriptApi
from pydantic import BaseModel
import marqo  # pyright: ignore[reportMissingTypeStubs]
import marqo.errors as marqo_errors  # pyright: ignore[reportMissingTypeStubs]
import dotenv
import os

def get_channel_playlist_id(channel_id: str, api_key: str) -> str | None:
    youtube = build('youtube', 'v3', developerKey=api_key)

    channel_response = youtube.channels().list(
        part='contentDetails',
        id=channel_id
    ).execute()

    return (
        channel_response
            .get('items', []).pop()
            .get('contentDetails', {})
            .get('relatedPlaylists', {})
            .get('uploads')
    )


def get_playlist_video_ids(playlist_id: str, api_key: str) -> list[str]:
    youtube = build('youtube', 'v3', developerKey=api_key)

    next_page_token: str = ""
    video_ids: list[str] = []

    while True:
        resp = youtube.playlistItems().list(
            part='contentDetails',
            playlistId=playlist_id,
            maxResults=50,
            pageToken=next_page_token
        ).execute()

        for item in resp.get('items', []):
            cur_id = item.get('contentDetails', {}).get('videoId')
            if cur_id is not None:
                video_ids += [cur_id]

        cur_npt = resp.get('nextPageToken')
        if cur_npt is not None:
            next_page_token = cur_npt
        else:
            break

    return video_ids

class VideoMetadata(BaseModel):
    video_id: str
    title: str
    thumbnail: str
    channel_id: str
    description: str
    upload_date: str
    url: str
    channel_name: str
    transcript: str | None = None

def get_videos_metadata(video_ids: list[str], api_key: str) -> list[VideoMetadata]:
    youtube = build('youtube', 'v3', developerKey=api_key)

    videos: list[VideoMetadata] = []
    for i in range(0, len(video_ids), 50):
        batch_ids = video_ids[i:i+50]
        resp = youtube.videos().list(
            part='snippet',
            id=','.join(batch_ids)
        ).execute()

        for item in resp.get('items', []):
            snippet = item.get('snippet', {})
            thumbnail = (
                snippet
                    .get('thumbnails', {})
                    .get('standard', {})
                    .get('url', "not found")
            )

            cur_vid_meta = VideoMetadata(
                video_id=item.get('id', "not found"),
                title=snippet.get('title', "not found"),
                thumbnail=thumbnail,
                channel_id=snippet.get('channelId', "not found"),
                description=snippet.get('description', "not found"),
                upload_date=snippet.get('publishedAt', "not found"),
                url=f"https://www.youtube.com/watch?v={item.get('id', 'not found')}",
                channel_name=snippet.get('channelTitle', "not found")
            )

            videos.append(cur_vid_meta)

    return videos

def get_existing_video_ids(db_client: marqo.Client, index_name: str) -> list[str]:
    try:
        search_results = db_client.index(index_name).search(  # pyright: ignore[reportUnknownMemberType]
            q="",
            limit=1000,
            searchable_attributes=["title"],
        )
    except marqo_errors.MarqoWebError as e:
        if e.status_code == 404:
            return [] # Index does not exist yet
        else:
            raise e

    ids: list[str] = []
    for hit in search_results.get('hits', []):  # pyright: ignore[reportAny]
        video_id = str(hit.get('video_id'))  # pyright: ignore[reportAny]
        ids.append(video_id)

    return ids

index_settings = {
    "type": "structured",
    "textPreprocessing": {
        "splitLength": 5,
        "splitOverlap": 1,
        "splitMethod": "sentence"
    },
    "allFields": [
        {"name": "video_id", "type": "text"},
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

# Throws error
def put_video(db_client: marqo.Client, index_name: str, video_meta: VideoMetadata, transcript: str | None) -> None:

    try:
        print(f"Checking if index {index_name} exists...")
        _ = db_client.get_index(index_name)
        print(f"Index {index_name} exists.")
    except marqo_errors.MarqoWebError as e:
        if e.status_code != 404:
            raise e
        else:
            print(f"Index {index_name} does not exist. Creating...")
            resp = db_client.create_index(index_name, settings_dict=index_settings)
            print(f"Created index {index_name}: {resp}")

    video_meta.transcript = transcript
    raw_video_meta = video_meta.model_dump()

    resp = db_client.index(index_name).add_documents([raw_video_meta])   # pyright: ignore[reportUnknownMemberType]

    if isinstance(resp, dict) and resp.get('errors', True) == True:
        raise Exception(f"Error indexing video {video_meta.video_id}: {resp}")
    
    print(f"Indexed video {video_meta.video_id}: {resp}")


def get_video_transcript(video_id: str) -> str | None:
    yttapi = YouTubeTranscriptApi()
    
    max_retries = 1 # after 20 mins of wait
    retry_count = 0
    transcript = None
    while retry_count < max_retries:
        try:
            transcript = yttapi.fetch(video_id)
        except Exception as e:
            if "blocking requests from your IP." in str(e):
                print(f"Transcript fetching blocked for video {video_id}: {e}")
                print(f"Retrying in 10 mins...")
                sleep(20 * 60)
            else:
                print(f"Error fetching transcript for video {video_id}: {e}")
                return None

    if transcript is None:
        print(f"Failed to fetch transcript for video {video_id} after {max_retries} retries.")
        return None

    transcript_text = ""
    for snippet in transcript:
        transcript_text += snippet.text

    return transcript_text

def refresh_videos_for_channel(channel_id: str, api_key: str, db_client: marqo.Client, index_name: str) -> None:
    playlist_id = get_channel_playlist_id(channel_id, api_key)
    if playlist_id is None:
        print("Could not retrieve playlist ID.")
        return

    video_ids = get_playlist_video_ids(playlist_id, api_key)
    print(f"Found {len(video_ids)} videos in channel {channel_id}.")

    existing_ids = get_existing_video_ids(db_client, index_name)
    print(f"Found {len(existing_ids)} existing videos in index {index_name}.")

    new_video_ids = [vid for vid in video_ids if vid not in existing_ids]
    print(f"Found {len(new_video_ids)} new videos to index.")

    if len(new_video_ids) == 0:
        print("No new videos to index. Exiting.")
        return

    videos_metadata = get_videos_metadata(new_video_ids, api_key)

    for vid_meta in videos_metadata:
        transcript = get_video_transcript(vid_meta.video_id)
        put_video(db_client, index_name, vid_meta, transcript)


def main():
    _ = dotenv.load_dotenv("../.env")

    index_name = os.getenv("MARQO_INDEX_NAME")
    if index_name is None:
        raise ValueError("MARQO_INDEX_NAME environment variable not set.")

    marqo_url = os.getenv("MARQO_HOST")
    if marqo_url is None:
        raise ValueError("MARQO_HOST environment variable not set.")

    channels = os.getenv("CHANNELS", "").split(",")
    if len(channels) == 0:
        raise ValueError("YOUTUBE_CHANNEL_IDS environment variable not set.")

    api_key = os.getenv("YOUTUBE_API_KEY")
    if api_key is None:
        raise ValueError("YOUTUBE_API_KEY environment variable not set.")

    db_client = marqo.Client(marqo_url)

    for channel_id in channels:
        print(f"Refreshing videos for channel {channel_id}...")
        refresh_videos_for_channel(channel_id, api_key, db_client, index_name)

if __name__ == "__main__":
    main()
