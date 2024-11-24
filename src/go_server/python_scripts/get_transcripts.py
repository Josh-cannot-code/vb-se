from youtube_transcript_api import YouTubeTranscriptApi
import argparse
import json

# TODO: failure cases
if __name__ == "__main__":
    parser = argparse.ArgumentParser(prog="dowload_video")
    parser.add_argument("videoIds")
    args = parser.parse_args()
    videoIds = args.videoIds.replace("'", "").split(",")

    transcripts, unretrievable_transcripts = YouTubeTranscriptApi.get_transcripts(videoIds, continue_after_error=True)

    transcriptMap = {} 
    for v_id in videoIds:
        if v_id in unretrievable_transcripts:
            #print(f"Unable to retrieve transcript for video: {v_id}")
            transcriptMap[v_id] = "NO_TRANSCRIPT"
        else:
            s = ""
            for t in transcripts[v_id]:
                s += " " + t["text"]
            transcriptMap[v_id] = s

    print(json.dumps(transcriptMap))
