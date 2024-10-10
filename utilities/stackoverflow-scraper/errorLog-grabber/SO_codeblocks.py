import redis
import re

# Connect to Redis
r = redis.Redis(host='logparse-nix.internal-headscale.ucaia.com', port=6379)

# Ping the server to ensure the connection works
try:
    r.ping()
    print("Connected to Redis server!")
except redis.ConnectionError:
    print("Failed to connect to Redis server.")

# Key name for the LIST in Redis stack server
list_key = "[nginx]"

list_length = r.llen(list_key)

# Change the range here for looking at different indexes. (E.G: looking at 0 index to 5)
list_elements = r.lrange(list_key, 0, 55237)

# Decode byte values (if they are stored as bytes)
decoded_list_elements = [element.decode('utf-8') for element in list_elements]

def extract_code_content(html):
    # Regex pattern to match content inside <code> tags
    code_pattern = re.compile(r'<code>(.*?)</code>', re.DOTALL)

    # Find all matches in the HTML string
    code_contents = code_pattern.findall(html)

    # Return a list of code content or concatenate them if needed
    return code_contents

# Initialize an empty list to hold all extracted logs
code_contents = []

# Hash key that contains stack overflow posts in Redis (replace with your actual hash key)
hash_key = "so_data"
for field in decoded_list_elements:
    # Retrieve the value from the hash for the specific field
    hash_value = r.hget(hash_key, field)
    # Check if the value exists in the hash
    if hash_value:
        # Decode the hash value from bytes to string
        decoded_value = hash_value.decode('utf-8')
        
        # Extract error logs from the decoded value
        code_content = extract_code_content(decoded_value)
        
        # Append the extracted logs to the main list
        code_contents.extend(code_content)
    else:
        print(f"Field '{field}' not found in hash '{hash_key}'")

# Save extracted code contents to a file for grep usage
output_file = 'code_contents.txt'

# Open the file in write mode and save the extracted contents
with open(output_file, 'w') as file:
    for content in code_contents:
        file.write(content + '\n')

print(f"Code contents saved to {output_file}")

#grep -E '[0-9]+#[0-9]+: \*.* open' code_contents.txt > output_logs.txt