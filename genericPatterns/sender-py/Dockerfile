FROM python:3.9-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENV OUTPUT_PIPES="kafka://your-kafka-broker;"

# Den Python-Server starten
CMD ["python", "main.py"]
