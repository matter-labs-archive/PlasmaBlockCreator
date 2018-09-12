const util = require("util");

async function getRedisFunctions(redisClient) {
    const redisExists = util.promisify(redisClient.exists).bind(redisClient);
    const redisGet = util.promisify(redisClient.get).bind(redisClient);
    const redisSet = util.promisify(redisClient.set).bind(redisClient);
    const redisIncr = util.promisify(redisClient.incr).bind(redisClient);
    const redisScript = util.promisify(redisClient.SCRIPT).bind(redisClient);
    const loadResult = await redisScript("LOAD", "local c = tonumber(redis.call('get', KEYS[1])); if c then if tonumber(ARGV[1]) > c then redis.call('set', KEYS[1], ARGV[1]) return tonumber(ARGV[1]) - c else return c - tonumber(ARGV[1]) end else return 0 end");
    console.log("Redis script sha = " + loadResult)
    const redisEvalSHA = util.promisify(redisClient.evalsha).bind(redisClient);
    const redisFunctions = {
        redisExists, redisGet, redisSet, redisIncr, redisScript, redisEvalSHA
    }
    return redisFunctions;
}

module.exports = getRedisFunctions;