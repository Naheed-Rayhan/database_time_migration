# MongoDB Time Migration Utility

This utility is designed to help you migrate time fields in your MongoDB collections from Bangladesh Time (BDT) to Coordinated Universal Time (UTC). It is particularly useful if you have stored timestamps in BDT and need to convert them to UTC for consistency or other reasons.


## Work that needs to be done
```
1. ModelTest -> 
2. LiveExam  ->  X
3. LiveClass
4. HomeWork
   i. AnimatedVideo
   ii. Quiz

```

## Fields that needs to be converted with their respective collections and databases
```
live-exam-dev
   model_tests
      {"exam_date", "result_publish_time", "created_at", "updated_at"}
   model_test_session
      {"start_time", "end_time", "created_at", "updated_at"}
   model_test_session_relation
      {"created_at"}
   model_tests_result_state
      {"created_at", "updated_at"}
      
      
academic-program-dev
   live_class
      {"start_time", "end_time", "created_at", "updated_at"}
   lessons -> {content_type:"HomeWork" ,content_sub_type :"AnimatedVideo"}
              {content_type:"HomeWork" ,content_sub_type :"Quiz"}
      {"start_time", "end_time", "created_at", "updated_at"}
      
```

## Features

- **Flexible Field Selection**: You can specify which fields in your documents need to be converted.
- **Field Existence Check**: The utility checks if the specified fields exist in the documents before attempting to update them.
- **Batch Processing**: You can limit the number of documents to update in a single run, which is useful for large collections.
- **Filtering**: You can apply filters to update only specific documents that match certain criteria.
- **Logging**: The utility provides detailed logging to help you track the progress and results of the migration.

## Prerequisites

- Go installed on your machine.
- MongoDB connection URI.
- `.env` file to store your MongoDB URI (optional but recommended).

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/Naheed-Rayhan/database_time_migration.git
   cd database_time_migration
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Create a `.env` file in the root directory and add your MongoDB URI:
   ```env
   MONGODB_URI=mongodb://your_mongodb_uri
   ```

## Usage

1. **Load Environment Variables**: The utility will automatically load the `.env` file if it exists. If not, it will log a message but continue execution.

2. **Connect to MongoDB**: The utility will connect to your MongoDB instance using the provided URI.

3. **Specify Collections and Fields**: In the `main.go` file, specify the collections and fields you want to process. You can also set a limit on the number of documents to update.

4. **Run the Utility**: Execute the program using:
   ```bash
   go run main.go
   ```

## Example

Here’s an example of how to use the utility to convert time fields in different collections:

```
func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    // Getting the MongoDB URI from the environment variable
    uri := os.Getenv("MONGODB_URI")
    if uri == "" {
        log.Fatal("Set your 'MONGODB_URI' environment variable.")
    }

    // Connecting to MongoDB
    client, err := Database.ConnectToMongoDB(uri)
    if err != nil {
        log.Fatal(err)
    }
    defer func() {
        err := Database.DisconnectMongoDB(client)
        if err != nil {
            log.Fatal(err)
        }
    }()

    // Example: Convert time fields in the 'model_tests' collection
    coll1 := Database.GetCollection(client, "live-exam-dev", "model_tests")
    fieldsToProcess := []string{"exam_date", "result_publish_time", "created_at", "updated_at"}
    Utils.MigrateBDT2UTC(coll1, 1, fieldsToProcess, bson.M{})
}
```

## Benefits

- **Time Zone Consistency**: Ensures all timestamps are in UTC, which is a standard practice for storing time data.
- **Automated Process**: Reduces the risk of human error by automating the conversion process.
- **Scalable**: Can handle large datasets with the ability to limit the number of updates per run.
- **Customizable**: Allows you to specify which fields and documents to update, providing flexibility based on your needs.

## Logging

The utility logs detailed information about the migration process, including:

- The number of documents matched and updated.
- Any fields that were skipped due to missing or invalid data.
- Errors encountered during the process.

## Conclusion

This MongoDB Time Migration Utility is a powerful tool for ensuring your time data is consistent and accurate. By converting BDT timestamps to UTC, you can avoid potential issues with time zone discrepancies and improve the reliability of your data.

For any issues or feature requests, please open an issue on the GitHub repository.


