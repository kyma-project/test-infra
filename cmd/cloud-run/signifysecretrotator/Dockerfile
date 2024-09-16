FROM python:3.12.5-alpine3.20

# Allow statements and log messages to immediately appear in the Knative logs
ENV PYTHONUNBUFFERED True

WORKDIR /app

COPY . .

RUN pip install --no-cache-dir --upgrade '.[signify-secret-rotator]' && \
	apk add --no-cache ca-certificates

CMD ["gunicorn", "--bind", "0.0.0.0:8080", "--workers", "4", "--timeout", "0", "signify-secret-rotator"]
