FROM python:3.12-slim
WORKDIR /app
COPY scripts/knowledge/rerank-server.py /app/rerank-server.py
RUN pip install --no-cache-dir fastapi uvicorn sentence-transformers
EXPOSE 8092
CMD ["python3", "/app/rerank-server.py", "--host", "0.0.0.0", "--port", "8092"]
