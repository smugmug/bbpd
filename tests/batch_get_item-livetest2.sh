curl -H "X-Amz-Target: DynamoDB_20120810.BatchGetItem" -X POST -d '{"RequestItems":{"test-godynamo-livetest":{"AttributesToGet":null,"ConsistentRead":false,"Keys":[{"TheHashKey":{"S":"AHashKey101"},"TheRangeKey":{"N":"101"}},{"TheHashKey":{"S":"AHashKey102"},"TheRangeKey":{"N":"102"}},{"TheHashKey":{"S":"AHashKey103"},"TheRangeKey":{"N":"103"}},{"TheHashKey":{"S":"AHashKey104"},"TheRangeKey":{"N":"104"}},{"TheHashKey":{"S":"AHashKey105"},"TheRangeKey":{"N":"105"}},{"TheHashKey":{"S":"AHashKey106"},"TheRangeKey":{"N":"106"}},{"TheHashKey":{"S":"AHashKey107"},"TheRangeKey":{"N":"107"}},{"TheHashKey":{"S":"AHashKey108"},"TheRangeKey":{"N":"108"}},{"TheHashKey":{"S":"AHashKey109"},"TheRangeKey":{"N":"109"}},{"TheHashKey":{"S":"AHashKey110"},"TheRangeKey":{"N":"110"}},{"TheHashKey":{"S":"AHashKey111"},"TheRangeKey":{"N":"111"}},{"TheHashKey":{"S":"AHashKey112"},"TheRangeKey":{"N":"112"}},{"TheHashKey":{"S":"AHashKey113"},"TheRangeKey":{"N":"113"}},{"TheHashKey":{"S":"AHashKey114"},"TheRangeKey":{"N":"114"}},{"TheHashKey":{"S":"AHashKey115"},"TheRangeKey":{"N":"115"}},{"TheHashKey":{"S":"AHashKey116"},"TheRangeKey":{"N":"116"}},{"TheHashKey":{"S":"AHashKey117"},"TheRangeKey":{"N":"117"}},{"TheHashKey":{"S":"AHashKey118"},"TheRangeKey":{"N":"118"}},{"TheHashKey":{"S":"AHashKey119"},"TheRangeKey":{"N":"119"}},{"TheHashKey":{"S":"AHashKey120"},"TheRangeKey":{"N":"120"}},{"TheHashKey":{"S":"AHashKey121"},"TheRangeKey":{"N":"121"}},{"TheHashKey":{"S":"AHashKey122"},"TheRangeKey":{"N":"122"}},{"TheHashKey":{"S":"AHashKey123"},"TheRangeKey":{"N":"123"}},{"TheHashKey":{"S":"AHashKey124"},"TheRangeKey":{"N":"124"}},{"TheHashKey":{"S":"AHashKey125"},"TheRangeKey":{"N":"125"}},{"TheHashKey":{"S":"AHashKey126"},"TheRangeKey":{"N":"126"}},{"TheHashKey":{"S":"AHashKey127"},"TheRangeKey":{"N":"127"}},{"TheHashKey":{"S":"AHashKey128"},"TheRangeKey":{"N":"128"}},{"TheHashKey":{"S":"AHashKey129"},"TheRangeKey":{"N":"129"}},{"TheHashKey":{"S":"AHashKey130"},"TheRangeKey":{"N":"130"}},{"TheHashKey":{"S":"AHashKey131"},"TheRangeKey":{"N":"131"}},{"TheHashKey":{"S":"AHashKey132"},"TheRangeKey":{"N":"132"}},{"TheHashKey":{"S":"AHashKey133"},"TheRangeKey":{"N":"133"}},{"TheHashKey":{"S":"AHashKey134"},"TheRangeKey":{"N":"134"}},{"TheHashKey":{"S":"AHashKey135"},"TheRangeKey":{"N":"135"}},{"TheHashKey":{"S":"AHashKey136"},"TheRangeKey":{"N":"136"}},{"TheHashKey":{"S":"AHashKey137"},"TheRangeKey":{"N":"137"}},{"TheHashKey":{"S":"AHashKey138"},"TheRangeKey":{"N":"138"}},{"TheHashKey":{"S":"AHashKey139"},"TheRangeKey":{"N":"139"}},{"TheHashKey":{"S":"AHashKey140"},"TheRangeKey":{"N":"140"}},{"TheHashKey":{"S":"AHashKey141"},"TheRangeKey":{"N":"141"}},{"TheHashKey":{"S":"AHashKey142"},"TheRangeKey":{"N":"142"}},{"TheHashKey":{"S":"AHashKey143"},"TheRangeKey":{"N":"143"}},{"TheHashKey":{"S":"AHashKey144"},"TheRangeKey":{"N":"144"}},{"TheHashKey":{"S":"AHashKey145"},"TheRangeKey":{"N":"145"}},{"TheHashKey":{"S":"AHashKey146"},"TheRangeKey":{"N":"146"}},{"TheHashKey":{"S":"AHashKey147"},"TheRangeKey":{"N":"147"}},{"TheHashKey":{"S":"AHashKey148"},"TheRangeKey":{"N":"148"}},{"TheHashKey":{"S":"AHashKey149"},"TheRangeKey":{"N":"149"}},{"TheHashKey":{"S":"AHashKey150"},"TheRangeKey":{"N":"150"}},{"TheHashKey":{"S":"AHashKey151"},"TheRangeKey":{"N":"151"}},{"TheHashKey":{"S":"AHashKey152"},"TheRangeKey":{"N":"152"}},{"TheHashKey":{"S":"AHashKey153"},"TheRangeKey":{"N":"153"}},{"TheHashKey":{"S":"AHashKey154"},"TheRangeKey":{"N":"154"}},{"TheHashKey":{"S":"AHashKey155"},"TheRangeKey":{"N":"155"}},{"TheHashKey":{"S":"AHashKey156"},"TheRangeKey":{"N":"156"}},{"TheHashKey":{"S":"AHashKey157"},"TheRangeKey":{"N":"157"}},{"TheHashKey":{"S":"AHashKey158"},"TheRangeKey":{"N":"158"}},{"TheHashKey":{"S":"AHashKey159"},"TheRangeKey":{"N":"159"}},{"TheHashKey":{"S":"AHashKey160"},"TheRangeKey":{"N":"160"}},{"TheHashKey":{"S":"AHashKey161"},"TheRangeKey":{"N":"161"}},{"TheHashKey":{"S":"AHashKey162"},"TheRangeKey":{"N":"162"}},{"TheHashKey":{"S":"AHashKey163"},"TheRangeKey":{"N":"163"}},{"TheHashKey":{"S":"AHashKey164"},"TheRangeKey":{"N":"164"}},{"TheHashKey":{"S":"AHashKey165"},"TheRangeKey":{"N":"165"}},{"TheHashKey":{"S":"AHashKey166"},"TheRangeKey":{"N":"166"}},{"TheHashKey":{"S":"AHashKey167"},"TheRangeKey":{"N":"167"}},{"TheHashKey":{"S":"AHashKey168"},"TheRangeKey":{"N":"168"}},{"TheHashKey":{"S":"AHashKey169"},"TheRangeKey":{"N":"169"}},{"TheHashKey":{"S":"AHashKey170"},"TheRangeKey":{"N":"170"}},{"TheHashKey":{"S":"AHashKey171"},"TheRangeKey":{"N":"171"}},{"TheHashKey":{"S":"AHashKey172"},"TheRangeKey":{"N":"172"}},{"TheHashKey":{"S":"AHashKey173"},"TheRangeKey":{"N":"173"}},{"TheHashKey":{"S":"AHashKey174"},"TheRangeKey":{"N":"174"}},{"TheHashKey":{"S":"AHashKey175"},"TheRangeKey":{"N":"175"}},{"TheHashKey":{"S":"AHashKey176"},"TheRangeKey":{"N":"176"}},{"TheHashKey":{"S":"AHashKey177"},"TheRangeKey":{"N":"177"}},{"TheHashKey":{"S":"AHashKey178"},"TheRangeKey":{"N":"178"}},{"TheHashKey":{"S":"AHashKey179"},"TheRangeKey":{"N":"179"}},{"TheHashKey":{"S":"AHashKey180"},"TheRangeKey":{"N":"180"}},{"TheHashKey":{"S":"AHashKey181"},"TheRangeKey":{"N":"181"}},{"TheHashKey":{"S":"AHashKey182"},"TheRangeKey":{"N":"182"}},{"TheHashKey":{"S":"AHashKey183"},"TheRangeKey":{"N":"183"}},{"TheHashKey":{"S":"AHashKey184"},"TheRangeKey":{"N":"184"}},{"TheHashKey":{"S":"AHashKey185"},"TheRangeKey":{"N":"185"}},{"TheHashKey":{"S":"AHashKey186"},"TheRangeKey":{"N":"186"}},{"TheHashKey":{"S":"AHashKey187"},"TheRangeKey":{"N":"187"}},{"TheHashKey":{"S":"AHashKey188"},"TheRangeKey":{"N":"188"}},{"TheHashKey":{"S":"AHashKey189"},"TheRangeKey":{"N":"189"}},{"TheHashKey":{"S":"AHashKey190"},"TheRangeKey":{"N":"190"}},{"TheHashKey":{"S":"AHashKey191"},"TheRangeKey":{"N":"191"}},{"TheHashKey":{"S":"AHashKey192"},"TheRangeKey":{"N":"192"}},{"TheHashKey":{"S":"AHashKey193"},"TheRangeKey":{"N":"193"}},{"TheHashKey":{"S":"AHashKey194"},"TheRangeKey":{"N":"194"}},{"TheHashKey":{"S":"AHashKey195"},"TheRangeKey":{"N":"195"}},{"TheHashKey":{"S":"AHashKey196"},"TheRangeKey":{"N":"196"}},{"TheHashKey":{"S":"AHashKey197"},"TheRangeKey":{"N":"197"}},{"TheHashKey":{"S":"AHashKey198"},"TheRangeKey":{"N":"198"}},{"TheHashKey":{"S":"AHashKey199"},"TheRangeKey":{"N":"199"}},{"TheHashKey":{"S":"AHashKey200"},"TheRangeKey":{"N":"200"}}]}},"ReturnConsumedCapacity":"NONE"}' http://localhost:12333/
