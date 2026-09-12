import asyncio
from dotenv import load_dotenv
import os
from openai import AsyncOpenAI

load_dotenv()

async def list_models():
    client = AsyncOpenAI(
        api_key=os.getenv("AI_API_KEY"),
        base_url=os.getenv("AI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta/openai/")
    )
    
    try:
        response = await client.models.list()
        print("Available models:")
        for model in response.data:
            print(f"- {model.id}")
    except Exception as e:
        print(f"Error listing models: {e}")

if __name__ == "__main__":
    asyncio.run(list_models())
