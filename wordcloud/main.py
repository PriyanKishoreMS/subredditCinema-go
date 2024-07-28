from wordcloud import WordCloud
from collections import defaultdict
import json
import sys
import random

def get_color_func(word=None, font_size=None, position=None, orientation=None, font_path=None, random_state=None):
    colors = [
        "#f97316",  
        "#0ea5e9",  
        "#eab308",  
        "#ef4444",  
        "#22c55e",  
        "#8b5cf6",  
        "#ec4899",  
    ]
    return random.choice(colors)

def json_to_multidict(json_data):
    multidict = defaultdict(int)
    try:
        data = json.loads(json_data)
        first_key = next(iter(data))
        word_list = data[first_key]
    
        for item in word_list:
            word = item.get("Word")
            count = item.get("Count")
            if word and count is not None:
                multidict[word] += count
    except json.JSONDecodeError as e:
        print(f"JSON Decode Error: {str(e)}", file=sys.stderr)
        raise
    except Exception as e:
        print(f"Error in json_to_multidict: {str(e)}", file=sys.stderr)
        raise
    return first_key, dict(multidict)

def makeImage(text, filename):
    try:
        wc = WordCloud(
            background_color="#10142C",
            scale=5,
            max_words=100,
            max_font_size=80,
            min_font_size=10,
            width=600,
            height=400,
            color_func=get_color_func,
            # font_path="/path/to/your/font.ttf",  # Replace with path to a custom font if desired
            prefer_horizontal=0.7,
            relative_scaling=0.5,
            margin=10,
            random_state=42
        )
        wc.generate_from_frequencies(text)
        wc.to_file(f"public/wordcloud/{filename}_wordcloud.png")
        print(f"Word cloud generated: {filename}_wordcloud.png", file=sys.stderr)
    except Exception as e:
        print(f"Error in makeImage: {str(e)}", file=sys.stderr)
        raise


if __name__ == "__main__":
    try:
        json_string = sys.stdin.read()
        key, result = json_to_multidict(json_string)
        makeImage(result, key)
        print(f"Generated wordcloud for {key}")
    except Exception as e:
        print(f"Main Error: {str(e)}", file=sys.stderr)
        sys.exit(1)