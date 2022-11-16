var databases = [
    {
        name: "files",
        collections: ["metadata", "collections"]
    }
];

for (database of databases) {
    temp = db.getSiblingDB(database.name);
    for (collection of database.collections){
        temp.createCollection(collection);
    }
    temp.createUser({ user: 'tester', pwd: 'testing', roles: [{role: 'dbOwner', db: database.name }, {role: 'read', db: 'admin' }] })
}
