import { app, InvocationContext, Timer, input } from "@azure/functions";
import { QueueServiceClient } from "@azure/storage-queue";

const connectionString = process.env.AZURE_STORAGE_CONNECTION_STRING;
const pollingTriggerQueueName = process.env.POLLING_TRIGGER_QUEUE_NAME;
const queueServiceClient = QueueServiceClient.fromConnectionString(connectionString);
const caDphPollingCron = process.env.CA_DPH_POLLING_CRON;

export async function caDphTimerTrigger(myTimer: Timer, context: InvocationContext): Promise<void> {

    const queueClient = queueServiceClient.getQueueClient(pollingTriggerQueueName)

    // We set the visibility timeout for the message on reading, in queue.go
    // messageTimeToLive of -1 means the message does not expire
    // the queue message contents will (in future) be the key to client-specific config
    const sendMessageResponse = await queueClient.sendMessage("cadph", {messageTimeToLive: -1})
    console.log("Sent message successfully, service assigned message Id:", sendMessageResponse.messageId, "service assigned request Id:", sendMessageResponse.requestId );

    context.log('Timer function processed request.');
}

app.timer('caDphTimerTrigger', {
    schedule: caDphPollingCron,
    handler: caDphTimerTrigger,
});
