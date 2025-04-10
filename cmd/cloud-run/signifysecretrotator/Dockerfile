FROM python:3.13.3-alpine3.21

# Allow statements and log messages to immediately appear in the Knative logs
ENV PYTHONUNBUFFERED True

WORKDIR /app

COPY ./cmd/cloud-run/signifysecretrotator .
COPY ./cmd/cloud-run/signifysecretrotator/requirements.txt .

# Install the required packages
RUN apk update \
	&& apk add gcc g++ linux-headers
RUN pip install --no-cache-dir  --upgrade -r requirements.txt && \
	apk add --no-cache ca-certificates

CMD ["gunicorn", "--bind", "0.0.0.0:8080", "--workers", "4", "--timeout", "0", "signifysecretrotator:app"]
