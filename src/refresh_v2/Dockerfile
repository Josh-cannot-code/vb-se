FROM python:3.12-slim-trixie
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

COPY . /app

WORKDIR /app
RUN uv sync --locked

ENV PYTHONUNBUFFERED=1

CMD ["uv", "run", "main.py"]
