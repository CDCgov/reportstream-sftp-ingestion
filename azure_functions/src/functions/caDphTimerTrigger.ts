import { app, InvocationContext, Timer } from "@azure/functions";
import { QueueServiceClient } from "@azure/storage-queue";

const connectionString = process.env.AZURE_STORAGE_CONNECTION_STRING;
const queueServiceClient = QueueServiceClient.fromConnectionString(connectionString);

export async function caDphTimerTrigger(myTimer: Timer, context: InvocationContext): Promise<void> {
    /* TODO -
        - Figure out TF for the Azure function
        - Make sure Azure Function has access to env vars
        - Figure out local testing
        - Figure out how to enqueue a message from here
        - Create a new queue for timer triggers - probably one total, and the message is the customer?
        - Create a DLQ for it
        - Create a queue reader including dead lettering
        - Profit?
    */

    // TODO - get queue name from env vars
    const queueClient = queueServiceClient.getQueueClient("queuename")

    // TODO - check on options for send message
    queueClient.sendMessage("cheezburger")
    context.log('Timer function processed request.');
}
// TODO - add info about installing typescript
// TODO - is .funcignore at the right level?
// TODO - set up the right CRON expression
// TODO - figure out how we make sure there's only one Azure Function running
// TODO - figure out if we can add multiple timers (like one per external customer?) to the same function
app.timer('caDphTimerTrigger', {
    schedule: '0 */1 * * * *',
    handler: caDphTimerTrigger
});
