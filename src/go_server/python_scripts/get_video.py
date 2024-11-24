import yt_dlp
import argparse
import sys
import os
import json

# TODO: Failure cases
if __name__ == "__main__":
    parser = argparse.ArgumentParser(prog="dowload_video")
    parser.add_argument("videoId")
    args = parser.parse_args()

    options = {
            "quiet": True,
    }

    videoId = args.videoId.replace("'", "")

    with yt_dlp.YoutubeDL(options) as ydl:
        result = ydl.extract_info('https://www.youtube.com/watch?v=' + videoId, download=False)

    video = {
        'id': result['id'],
        'title': result['title'],
        'thumbnail': result['thumbnail'],
        'channel_id': result['channel_id'],
        'description': result['description'],
        'upload_date': result['upload_date'],
        'url': result['original_url'],
        'channel_name': result['channel']
    }

    json_video = json.dumps(video)

    print(json_video)
