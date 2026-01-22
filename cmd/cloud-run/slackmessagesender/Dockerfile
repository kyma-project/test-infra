# Stage 1: Build with dependencies inside venv
FROM python:3.14.2-alpine3.23 AS builder

ENV PYTHONUNBUFFERED=1

WORKDIR /app

# Create a virtual environment
RUN python -m venv /app/venv

# Activate venv and install dependencies
ENV PATH="/app/venv/bin:$PATH"

COPY ./cmd/cloud-run/slackmessagesender/requirements.txt .

RUN pip install --upgrade pip && pip install --no-cache-dir -r requirements.txt

# Stage 2: Production image
FROM python:3.14.2-alpine3.23

WORKDIR /app

ENV PYTHONUNBUFFERED=1

# Copy app source code
COPY ./cmd/cloud-run/slackmessagesender/slackmessagesender.py .

# Copy the virtual environment from the builder stage
COPY --from=builder /app/venv /app/venv
ENV PATH="/app/venv/bin:$PATH"

ENTRYPOINT ["/app/venv/bin/python"]
CMD ["/app/venv/bin/gunicorn", "--bind", "0.0.0.0:8080", "--workers", "4", "--timeout", "0", "slackmessagesender:app"]
