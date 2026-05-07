const mongoHost = process.env.AMBULANCE_API_MONGODB_HOST;
const mongoPort = process.env.AMBULANCE_API_MONGODB_PORT;

const mongoUser = process.env.AMBULANCE_API_MONGODB_USERNAME;
const mongoPassword = process.env.AMBULANCE_API_MONGODB_PASSWORD;

const database = process.env.AMBULANCE_API_MONGODB_DATABASE;
const collection = process.env.AMBULANCE_API_MONGODB_COLLECTION;
const assignmentCollection =
  process.env.AMBULANCE_API_MONGODB_ASSIGNMENT_COLLECTION || "department_assignments";

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

// if database and BOTH collections exist, exit with success - already initialized
const databases = connection.getDBNames();
if (databases.includes(database)) {
  const dbInstance = connection.getDB(database);
  const existing = dbInstance.getCollectionNames();
  if (existing.includes(collection) && existing.includes(assignmentCollection)) {
    print(
      `Collections '${collection}' and '${assignmentCollection}' already exist in database '${database}'`,
    );
    process.exit(0);
  }
}

// initialize
const db = connection.getDB(database);
const existingCollections = db.getCollectionNames();
const performanceJustCreated = !existingCollections.includes(collection);
const assignmentJustCreated = !existingCollections.includes(assignmentCollection);

if (performanceJustCreated) {
  db.createCollection(collection);
  // create indexes
  db[collection].createIndex({ id: 1 }, { unique: true });
  db[collection].createIndex({ employeeId: 1 });
  db[collection].createIndex({ date: 1 });
}

if (assignmentJustCreated) {
  db.createCollection(assignmentCollection);
  db[assignmentCollection].createIndex({ id: 1 }, { unique: true });
  db[assignmentCollection].createIndex({ employeeId: 1 });
  db[assignmentCollection].createIndex({ departmentId: 1 });
}

// insert sample performance records only on first creation
// (matches the PerformanceRecord schema in api/vykon.openapi.yaml)
// IDs use NumberLong so they are stored as BSON Long, matching the Go int64 type used by the API.
let result = performanceJustCreated ? db[collection].insertMany([
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
]) : null;

if (result && result.writeError) {
  console.error(result);
  print(`Error when writing the data: ${result.errmsg}`);
}

// insert sample department assignments only on first creation
// (matches DepartmentAssignment schema in api/assignment.openapi.yaml)
let assignmentResult = assignmentJustCreated ? db[assignmentCollection].insertMany([
  {
    id: NumberLong(1),
    employeeId: "EMP-001",
    employeeName: "MUDr. Jana Nováková",
    departmentId: "DEPT-CARD",
    departmentName: "Kardiologická ambulancia",
    role: "Vedúci lekár",
    fromDate: "2026-01-01",
    toDate: "2026-12-31",
    note: "Rotácia v rámci internej kliniky",
  },
  {
    id: NumberLong(2),
    employeeId: "EMP-002",
    employeeName: "MUDr. Peter Kováč",
    departmentId: "DEPT-SURG",
    departmentName: "Chirurgické oddelenie",
    role: "Lekár",
    fromDate: "2026-03-15",
  },
  {
    id: NumberLong(3),
    employeeId: "EMP-003",
    employeeName: "Mgr. Eva Horváthová",
    departmentId: "DEPT-ER",
    departmentName: "Pohotovosť",
    role: "Sestra",
    fromDate: "2026-02-01",
    toDate: "2026-04-30",
  },
]) : null;

if (assignmentResult && assignmentResult.writeError) {
  console.error(assignmentResult);
  print(`Error when writing the assignment data: ${assignmentResult.errmsg}`);
}

// exit with success
process.exit(0);
