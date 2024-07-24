import { app, InvocationContext, Timer, input } from "@azure/functions";
import { QueueServiceClient } from "@azure/storage-queue";

const connectionString = process.env.AZURE_STORAGE_CONNECTION_STRING;
const pollingTriggerQueueName = process.env.POLLING_TRIGGER_QUEUE_NAME;
const queueServiceClient = QueueServiceClient.fromConnectionString(connectionString);

export async function caDphTimerTrigger(myTimer: Timer, context: InvocationContext): Promise<void> {

    console.log(connectionString)
    const queueClient = queueServiceClient.getQueueClient(pollingTriggerQueueName)
    console.log("Timer:")
    console.log(myTimer);
    console.log("Context:")
    console.log(context);
    // context.extraInputs.get("customer")

    // We set the visibility timeout for the message on reading, in queue.go
    // messageTimeToLive of -1 means the message does not expire
    // the queue message contents will (in future) be the key to client-specific config
    const sendMessageResponse = await queueClient.sendMessage("cadph", {messageTimeToLive: -1})
    console.log("Sent message successfully, service assigned message Id:", sendMessageResponse.messageId, "service assigned request Id:", sendMessageResponse.requestId );

    context.log('Timer function processed request.');
}
// TODO - set up the right CRON expression
// TODO - figure out if we can add multiple timers (like one per external customer?) to the same function
// TODO - find out the timer's timezone for scheduling - there's a `schedule: { adjustForDST: true }` setting in the timer
// TODO - Azure SDK Timeout setup research for when calling azure doesn't work.
// https://learn.microsoft.com/en-us/azure/azure-functions/functions-bindings-timer?tabs=python-v2%2Cisolated-process%2Cnodejs-v4&pivots=programming-language-typescript#ncrontab-time-zones
app.timer('caDphTimerTrigger', {
    schedule: '0 30 9 * Jan Mon', // 9:30 am every monday in january
    handler: caDphTimerTrigger,
    // I don't think extraInputs is the right field, just messing around looking for
    // whether we can pass in a variable - then we could use one handler for every
    // customer that has a timer
    // Possibly the name (in the first param above) could work for this?
    // const customer = input.generic({"type": "customer", "name": "CADPH"})
    // extraInputs: [customer]
});
