# Stage 1: Build with dependencies inside venv
FROM cgr.dev/chainguard/python:latest-dev as builder

ENV PYTHONUNBUFFERED=1

WORKDIR /app

# Create a virtual environment
RUN python -m venv /app/venv

# Activate venv and install dependencies
ENV PATH="/app/venv/bin:$PATH"

COPY ./cmd/cloud-run/slack-message-sender/requirements.txt .

RUN pip install --upgrade pip && pip install --no-cache-dir -r requirements.txt

# Stage 2: Production image
FROM cgr.dev/chainguard/python:latest

WORKDIR /app

ENV PYTHONUNBUFFERED=1

# Copy app source code
COPY ./cmd/cloud-run/slack-message-sender/slack-message-sender.py .

# Copy the virtual environment from the builder stage
COPY --from=builder /app/venv /app/venv
ENV PATH="/app/venv/bin:$PATH"

CMD ["/app/venv/bin/gunicorn", "--bind", "0.0.0.0:8080", "--workers", "4", "--timeout", "0", "slack-message-sender:app"]
