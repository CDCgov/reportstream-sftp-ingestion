import { app, InvocationContext, Timer } from "@azure/functions";
import { QueueServiceClient } from "@azure/storage-queue";

const connectionString = process.env.AZURE_STORAGE_CONNECTION_STRING;
// TODO - add env vars to docker compose?
const pollingTriggerQueueName = process.env.POLLING_TRIGGER_QUEUE_NAME;
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

    const queueClient = queueServiceClient.getQueueClient(pollingTriggerQueueName)
    console.log(myTimer);
    console.log(context);
    // context.extraInputs.get("customer")
    // TODO - check on options for send message
    await queueClient.sendMessage("cheezburger")
    context.log('Timer function processed request.');
}
// TODO - add info about installing typescript
// TODO - figure out instructions for running this
// TODO - is .funcignore at the right level?
// TODO - set up the right CRON expression
// TODO - figure out how we make sure there's only one Azure Function running - we should alarm if there's more than one
// TODO - figure out if we can add multiple timers (like one per external customer?) to the same function
// TODO - find out the timer's timezone for scheduling
app.timer('caDphTimerTrigger', {
    schedule: '0 */1 * * * *',
    handler: caDphTimerTrigger,
    // I don't think extraInputs is the right field, just messing around looking for
    // whether we can pass in a variable - then we could use one handler for every
    // customer that has a timer
    // Possibly the name (in the first param above) could work for this?
    // extraInputs: [{name: "customer", type: ""}]
});
