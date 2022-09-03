import { Client } from 'discord.js';
import { createDailyGreeting } from './jobs/dailyGreeting';
import { schedule } from 'node-cron';
import { Messages } from '../../messages/messages';
import { createDailyReportReminder } from './jobs/dailyReport/dailyReportReminder';

interface InitSchedulerParams {
  client: Client<true>;
  messages: Messages;
  greetingChannelId: string;
  dailyReportChannelId: string;
  dailyReportTargetUserId: string;
}

export function initScheduler({
  client,
  messages,
  dailyReportTargetUserId,
  dailyReportChannelId,
  greetingChannelId,
}: InitSchedulerParams) {
  const jobs = [
    createDailyGreeting(greetingChannelId),
    createDailyReportReminder(
      dailyReportChannelId,
      dailyReportTargetUserId,
      messages.dailyReportReminder.afternoon,
      '00 16 * * *'
    ),
    createDailyReportReminder(
      dailyReportChannelId,
      dailyReportTargetUserId,
      messages.dailyReportReminder.night,
      '00 23 * * *'
    ),
  ];

  jobs.forEach(job => {
    schedule(
      job.cron,
      date => {
        console.log(
          `Running scheduled job ${job.name} at ${date.toISOString()}`
        );

        job
          .execute({
            date,
            messages,
            client,
          })
          .catch(error => {
            console.error(
              `Scheduled job ${job.name} failed at ${date.toISOString()}:`,
              error
            );
          });
      },
      {
        timezone: 'Europe/Warsaw',
      }
    );
  });

  console.log(`Scheduled ${jobs.length} jobs`);
}
