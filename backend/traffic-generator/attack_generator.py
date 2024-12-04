import requests
import time
import random
import os
from datetime import datetime
import urllib.parse

# Configuration
NGINX_HOST = os.getenv('NGINX_HOST', 'nginx')
NGINX_PORT = os.getenv('NGINX_PORT', '80')
REQUEST_INTERVAL = float(os.getenv('REQUEST_INTERVAL', '0.1'))
ATTACK_RATE = float(os.getenv('ATTACK_RATE', '0.8'))

# Common attack patterns
ATTACK_PATTERNS = [
    # Directory traversal attempts
    '/cgi-bin/.%2e/.%2e/.%2e/.%2e/.%2e/.%2e/.%2e/.%2e/.%2e/.%2e/bin/sh',
    '/cgi-bin/%%32%65%%32%65/%%32%65%%32%65/%%32%65%%32%65/%%32%65%%32%65/bin/sh',
    
    # Environment file probing
    '/.env',
    '/cms/.env',
    '/console/.env',
    '/admin/.env',
    '/.git/.env',
    '/.twilio.env',
    
    # PHPUnit vulnerabilities
    '/vendor/phpunit/phpunit/src/Util/PHP/eval-stdin.php',
    '/vendor/phpunit/Util/PHP/eval-stdin.php',
    '/lib/phpunit/phpunit/src/Util/PHP/eval-stdin.php',
    
    # Common sensitive paths
    '/wp-login.php',
    '/.git/config',
    '/robots.txt',
    '/favicon.ico'
]

# User agents from your logs
USER_AGENTS = [
    "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0; FunWebProducts)",
    "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; ru) Opera 8.01",
    "Mozilla/5.0 (Linux; Android 6.0.1; SAMSUNG SM-G550T1 Build/MMB29K) AppleWebKit/537.36",
    "Custom-AsyncHttpClient",
    "Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.8.1.11) Gecko/20071127 Firefox/2.0.0.11"
]

def generate_malicious_traffic():
    base_url = f'http://{NGINX_HOST}:{NGINX_PORT}'
    
    while True:
        try:
            if random.random() < ATTACK_RATE:
                # Generate attack traffic
                path = random.choice(ATTACK_PATTERNS)
                user_agent = random.choice(USER_AGENTS)
                
                if random.random() < 0.3:
                    # Sometimes add PHP/SQL injection patterns
                    path += f"?{urllib.parse.quote('?%ADd+allow_url_include%3d1+%ADd+auto_prepend_file%3dphp://input')}"
                
                headers = {
                    'User-Agent': user_agent,
                    'Accept': '*/*'
                }
                
                if random.random() < 0.2:
                    # Sometimes try POST requests
                    requests.post(f'{base_url}{path}', headers=headers, timeout=1)
                else:
                    requests.get(f'{base_url}{path}', headers=headers, timeout=1)
            else:
                # Generate some normal traffic
                path = '/'
                user_agent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
                headers = {'User-Agent': user_agent}
                requests.get(f'{base_url}{path}', headers=headers, timeout=1)

        except requests.exceptions.RequestException as e:
            print(f"Error making request: {e}")
            
        time.sleep(REQUEST_INTERVAL)

if __name__ == "__main__":
    print(f"Starting attack traffic generator targeting {NGINX_HOST}:{NGINX_PORT}")
    generate_malicious_traffic()
