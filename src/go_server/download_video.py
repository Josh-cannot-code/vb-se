import yt_dlp
import argparse
import sys
import os
import json

if __name__ == "__main__":
    # TODO: argparse
    videoId = "IELMSD2kdmk"
    options = {
            "quiet": True,
    }

    sys.stderr = open(os.devnull, "w")
    with yt_dlp.YoutubeDL(options) as ydl:
        result = ydl.extract_info('https://www.youtube.com/watch?v=' + videoId, download=False)
    sys.stderr = sys.__stdout__

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
