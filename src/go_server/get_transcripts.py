from youtube_transcript_api import YouTubeTranscriptApi
import argparse
import json

if __name__ == "__main__":
    videoIds = ["IELMSD2kdmk"]
    transcripts, unretrievable_transcripts = YouTubeTranscriptApi.get_transcripts(videoIds, continue_after_error=True)

    transcriptMap = {} 
    for v_id in videoIds:
        if v_id in unretrievable_transcripts:
            print(f"Unable to retrieve transcript for video: {v_id}")
            transcriptMap[v_id] = "NO_TRANSCRIPT"
            #data = { "transcript" : "NO TRANSCRIPT", "video_id" : v_id }
        else:
            s = ""
            for t in transcripts[v_id]:
                s += " " + t["text"]
            transcriptMap[v_id] = s
            #data = { "transcript" : s, "video_id" : v_id }

        print(json.dumps(transcriptMap))
