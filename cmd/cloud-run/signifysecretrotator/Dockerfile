# Stage 1: Build stage
FROM python:3.13.3-alpine3.21 AS builder

WORKDIR /app

COPY ./cmd/cloud-run/signifysecretrotator .
COPY ./cmd/cloud-run/signifysecretrotator/requirements.txt .

# Install build tools and Python dependencies
RUN apk update \
 && apk add --no-cache gcc g++ linux-headers \
 && pip install --no-cache-dir --upgrade -r requirements.txt

# Stage 2: Final stage
FROM python:3.13.3-alpine3.21

WORKDIR /app

# Copy application files
COPY ./cmd/cloud-run/signifysecretrotator .

# Copy installed Python dependencies from the builder stage
COPY --from=builder /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# Install runtime dependencies
RUN apk add --no-cache ca-certificates && \
    apk del sqlite

CMD ["gunicorn", "--bind", "0.0.0.0:8080", "--workers", "4", "--timeout", "0", "signifysecretrotator:app"]