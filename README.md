# go-jev

We kept running into scam bots in a discord server I'm in, and we thought it would be a cool use case for using Jev for auto-moderation features

Main import path is a `jev` client package, initial test in a static input is under `cmd/jev/`

Input:
```json
{
    "state": "@everyone Hey guys! I'm offering out my CANON EOS 1500D, a 18-55mm, a 17-85mm, a 50mm f/1.8, and a Kogan Horizon camera for free. They're in original box, nearly brand new condition. I bought them Christmas 2025, used only for test shoots. I'm offering it out because I just got a drone and want someone who needs it to have it. If you're interested, Text me on +[PHONE]",
    "model": "jev-latest",
    "questions": {
        "is_scam": {
            "type": "noul",
            "instructions": "Is this message likely a scam?"
        },
        "scam_level": {
            "type": "score",
            "instructions": "Rate the likelihood that this message is a scam.",
            "criteria": [
                "Low",
                "Medium",
                "High"
            ]
        }
    }
}
```

Sample output (do note output *is* non-deterministic):
```
Tokens Used: Input=433, Output=37

Answers:
  - [is_scam] Type: noul
    Noul Probability: 0.7300

  - [scam_level] Type: score
    Weighted Score: 1.68
    Confidence: 0.52
    Probabilities:
      * 1 (Medium): 0.2300 (23.0%)
      * 2 (High): 0.7300 (73.0%)
      * 0 (Low): 0.0400 (4.0%)
```
