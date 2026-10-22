FROM python:3.10-slim
WORKDIR /app
COPY ai-agent/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY ai-agent/main.py .
COPY features.txt .
EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]