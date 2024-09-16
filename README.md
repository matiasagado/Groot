# Log Parse PoC

### What is this

This is a proof of concept for our log parsing AI efforts.

The current state of this is:

1. A mock logger services publishes log lines to the Redis queue.
2. Redis holds these log lines until the ai-core can consume them.
3. The AI-Core when logs are available will pop a log off the queue one-by-one, processing each one until fetching the next one.

### What is missing from this repo:

- After much experimentation we are using [TabbyAPI](https://github.com/theroyallab/tabbyAPI) as the interface to our LLM. The engine it is using [ExLLamaV2](https://github.com/turboderp/exllamav2) to run the model.
- The model we are using is a quantized version of Mixtral-instruct-8x7B. The model is on Huggingface as [Mixtral-8x7B-instruct-exl2](huggingface.co/turboderp/Mixtral-8x7B-instruct-exl2). It fits on the 2x RTX 3090 GPUs we have on MonsterMax.
- A lot of the https://vector.dev configurations are not checked in here. There has been a lot of work done on those transformations.

- Currently the database output needs to be built out. I'll update this when I complete that.


### Instructions

1. Connect to our head/tailscale network (ask Erik / @StealthBadger747 for help)
2. Create a Python venv for this project and install the `requirements.txt` file.
3. Add the `TABBY_API_KEY` to the `docker-compose.yml`
4. Now you can do `docker compose up --build` (the `--build` flag is important, otherwise changes will not be picked up).
