# pow
Proof Of Work example

Design and implement "Word of Wisdom" tcp server.
 • TCP server should be protected from DDOS attacks with the Proof of Work (https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should be used.
 • The choice of the POW algorithm should be explained.
 • After Proof Of Work verification, server should send one of the quotes from "word of wisdom" book or any other collection of the quotes.
 • Docker file should be provided both for the server and for the client that solves the POW challenge

# Word of Wisdom TCP Server with POW

## How to run

git clone {repo_url}

cd {cloned_repo_dir}

docker-compose up

## Proof of Work Algorithms

### 1. Hashcash POW
- **Description**: Client must find a nonce that produces a hash with a certain number of leading zeros
- **Pros**:
  - Simple to implement and understand
  - Well-tested and widely used
  - Suitable for basic DDoS protection
- **Cons**:
  - Vulnerable to ASIC mining
  - Fixed difficulty may not adapt well to varying network conditions

### 2. Scrypt POW
- **Description**: Uses memory-hard function requiring significant RAM
- **Pros**:
  - More resistant to ASIC mining
  - Better security against hardware acceleration
- **Cons**:
  - More complex to implement
  - Higher resource requirements
  - May be too heavy for simple DDoS protection

### 3. Equihash POW
- **Description**: Based on the Generalized Birthday Problem
- **Pros**:
  - Strong ASIC resistance
  - Memory-bound algorithm
- **Cons**:
  - Complex implementation
  - Higher verification time
  - Overkill for basic DDoS protection

## Our Choice: Hashcash POW

For this example, we've chosen to implement Hashcash POW algorithm. While it's not the most ASIC-resistant option, it provides a good balance of:
- Simplicity of implementation
- Effectiveness for basic DDoS protection
- Well-understood security properties
- Low resource requirements
