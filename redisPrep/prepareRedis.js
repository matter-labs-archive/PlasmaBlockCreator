const getRedisFunctions = require("./createRedis");
const redis = require('redis');

async function prepareRedis() {
	const redisClient = redis.createClient({host: "redis",
                                            port: 6379,
                                            string_numbers: true,
                                            password: null});
	const redisFunctions = await getRedisFunctions(redisClient);
    const {redisExists, redisSet} = redisFunctions
    const exists = await redisExists("ctr")
    if (!exists) {
        await redisSet("ctr", "4294967295"); // 4294967296 - 1
    }
	redisClient.quit();
	console.log("Done")
}

prepareRedis().then();