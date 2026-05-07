const mongoHost = process.env.AMBULANCE_API_MONGODB_HOST;
const mongoPort = process.env.AMBULANCE_API_MONGODB_PORT;

const mongoUser = process.env.AMBULANCE_API_MONGODB_USERNAME;
const mongoPassword = process.env.AMBULANCE_API_MONGODB_PASSWORD;

const database = process.env.AMBULANCE_API_MONGODB_DATABASE;
const collection = process.env.AMBULANCE_API_MONGODB_COLLECTION;

const retrySeconds = parseInt(process.env.RETRY_CONNECTION_SECONDS || "5") || 5;

// try to connect to mongoDB until it is not available
let connection;
while (true) {
  try {
    connection = Mongo(`mongodb://${mongoUser}:${mongoPassword}@${mongoHost}:${mongoPort}`);
    break;
  } catch (exception) {
    print(`Cannot connect to mongoDB: ${exception}`);
    print(`Will retry after ${retrySeconds} seconds`);
    sleep(retrySeconds * 1000);
  }
}

// if database and collection exists, exit with success - already initialized
const databases = connection.getDBNames();
if (databases.includes(database)) {
  const dbInstance = connection.getDB(database);
  collections = dbInstance.getCollectionNames();
  if (collections.includes(collection)) {
    print(`Collection '${collection}' already exists in database '${database}'`);
    process.exit(0);
  }
}

// initialize
// create database and collection
const db = connection.getDB(database);
db.createCollection(collection);

// create indexes
db[collection].createIndex({ id: 1 }, { unique: true });
db[collection].createIndex({ employeeId: 1 });
db[collection].createIndex({ date: 1 });

// insert sample performance records (matches the PerformanceRecord schema in api/vykon.openapi.yaml)
// IDs use NumberLong so they are stored as BSON Long, matching the Go int64 type used by the API.
let result = db[collection].insertMany([
  {
    id: NumberLong(1),
    employeeId: "EMP-001",
    employeeName: "MUDr. Jana Nováková",
    date: "2026-04-28",
    hoursWorked: 8,
    examinationCount: 12,
    operationCount: 1,
    shiftCount: 0,
    note: "Bežná ambulantná zmena",
  },
  {
    id: NumberLong(2),
    employeeId: "EMP-002",
    employeeName: "MUDr. Peter Kováč",
    date: "2026-04-29",
    hoursWorked: 12,
    examinationCount: 4,
    operationCount: 3,
    shiftCount: 1,
  },
  {
    id: NumberLong(3),
    employeeId: "EMP-003",
    employeeName: "Mgr. Eva Horváthová",
    date: "2026-04-30",
    hoursWorked: 6,
    examinationCount: 9,
    operationCount: 0,
    shiftCount: 0,
    note: "Skrátená zmena",
  },
]);

if (result.writeError) {
  console.error(result);
  print(`Error when writing the data: ${result.errmsg}`);
}

// exit with success
process.exit(0);
