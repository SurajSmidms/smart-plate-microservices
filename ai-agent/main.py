import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import google.generativeai as genai

app = FastAPI(title="Restaurant AI Agent Service", version="1.0")

api_key = os.getenv("GEMINI_API_KEY", "")
genai.configure(api_key=api_key)

# Load features.txt behavior specification
features_prompt = ""
if os.path.exists("features.txt"):
    with open("features.txt", "r") as f:
        features_prompt = f.read()

class ChatRequest(BaseModel):
    role: str
    prompt: str

@app.post("/chat")
def chat_with_agent(req: ChatRequest):
    try:
        model = genai.GenerativeModel(
            model_name="gemini-3.8-flash",
            system_instruction=f"You are an intelligent global restaurant management assistant. Follow these specifications strictly:\n{features_prompt}"
        )
        full_query = f"User Role: {req.role}\nRequest: {req.prompt}"
        response = model.generate_content(full_query)
        return {"ai_response": response.text}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))