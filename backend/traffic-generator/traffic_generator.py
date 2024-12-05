import requests
import time
import random
import os
from datetime import datetime

# Configuration from environment variables
NGINX_HOST = os.getenv('NGINX_HOST', 'nginx')
NGINX_PORT = os.getenv('NGINX_PORT', '80')
REQUEST_INTERVAL = float(os.getenv('REQUEST_INTERVAL', '1'))
ERROR_RATE = float(os.getenv('ERROR_RATE', '0.2'))

# List of endpoints to randomly access
ENDPOINTS = [
    '/',
    '/api/users',
    '/api/products',
    '/api/orders',
    '/api/settings',
    '/nonexistent',
    '/api/invalid',
]

# List of user agents to randomize requests
USER_AGENTS = [
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
    'Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)',
    'Mozilla/5.0 (Linux; Android 11; SM-G991B)',
]

def generate_traffic():
    base_url = f'http://{NGINX_HOST}:{NGINX_PORT}'
    
    while True:
        endpoint = random.choice(ENDPOINTS)
        user_agent = random.choice(USER_AGENTS)
        headers = {'User-Agent': user_agent}

        try:
            # Randomly decide to send malformed requests
            if random.random() < ERROR_RATE:
                # Generate various types of problematic requests
                if random.random() < 0.5:
                    # Malformed headers
                    headers['Content-Length'] = 'invalid'
                else:
                    # Invalid HTTP method
                    requests.request('INVALID', f'{base_url}{endpoint}', headers=headers)
            else:
                # Normal GET request
                requests.get(f'{base_url}{endpoint}', headers=headers)

        except requests.exceptions.RequestException as e:
            print(f"Error making request: {e}")

        time.sleep(REQUEST_INTERVAL)

if __name__ == "__main__":
    print(f"Starting traffic generator targeting {NGINX_HOST}:{NGINX_PORT}")
    generate_traffic()
