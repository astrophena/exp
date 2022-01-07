FROM python:alpine

# Copy source code to the container.
WORKDIR /usr/src/app
COPY . /usr/src/app

# Install Python dependencies.
RUN pip install --no-cache-dir -r requirements.txt

# Run as a separate user.
RUN addgroup -S astroflask && adduser -S astroflask -G astroflask
USER astroflask

CMD ["gunicorn", "app:app", "-b", "0.0.0.0:8080"]
