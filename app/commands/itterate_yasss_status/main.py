from pymongo import MongoClient
import pandas as pd
import json

def extract_score_ids():
    # Connect to MongoDB
    client = MongoClient('mongodb://localhost:27017')
    db = client['yasss_om-api_com']

    # Execute the query
    query = {
        "status_response_json.result.status": {"$eq": 1}
    }

    results = list(db.api_cache.find(query))
    print(results)

    # Extract score_ids
    score_ids = []
    for doc in results:
        try:
            # Assuming status_response_json is stored as a string and needs to be parsed
            if isinstance(doc['status_response_json'], str):
                response_data = json.loads(doc['status_response_json'])
            else:
                response_data = doc['status_response_json']

            # Extract score_id (adjust the path based on your actual JSON structure)
            if 'score_id' in doc:
                score_ids.append({'score_id': doc['score_id']})

        except (KeyError, json.JSONDecodeError) as e:
            print(f"Error processing document: {e}")

    # Create DataFrame and save to CSV
    df = pd.DataFrame(score_ids)
    csv_filename = 'score_ids.csv'
    df.to_csv(csv_filename, index=False)
    print(f"CSV file '{csv_filename}' has been created with {len(score_ids)} records")

    # Close the connection
    client.close()

if __name__ == "__main__":
    extract_score_ids()
