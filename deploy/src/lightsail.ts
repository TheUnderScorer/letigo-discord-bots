import * as pulumi from '@pulumi/pulumi';
import * as aws from '@pulumi/aws';
import * as awsx from '@pulumi/awsx';
import { projectRoot } from './data';
import * as path from 'path';

const stack = pulumi.getStack();
const name = pulumi.getProject();
const serviceName = `${stack}-${name}`;

const ecr = new aws.ecr.Repository(`${serviceName}-server-ecr`, {
  imageScanningConfiguration: {
    scanOnPush: false,
  },
});

const server = new awsx.ecr.Image(`${serviceName}-server`, {
  path: projectRoot,
  dockerfile: path.join(projectRoot, 'docker/server/Dockerfile'),
  repositoryUrl: ecr.repositoryUrl,
  env: {
    DOCKER_BUILDKIT: '1',
  },
});

export const repoName = ecr.name;
export const repoUrl = ecr.repositoryUrl;
export const serverImageUri = server.imageUri;

const containerService = new aws.lightsail.ContainerService(serviceName, {
  isDisabled: false,
  power: 'nano',
  scale: 1,
  name: serviceName,
});

const port = 3000;
const containerEnvKeys = [
  'BOT_TOKEN',
  'APP_ID',
  'PUBLIC_KEY',
  'DAILY_REPORT_CHANNEL_ID',
  'DAILY_REPORT_TARGET_USER_ID',
  'GUILD_ID',
  'GREETING_CHANNEL_ID',
];
const containerEnv = containerEnvKeys.reduce((acc, key) => {
  return {
    ...acc,
    [key]: process.env[key] as string,
  };
}, {});
const containerServiceDeployment =
  new aws.lightsail.ContainerServiceDeploymentVersion(serviceName, {
    publicEndpoint: {
      containerName: ecr.name,
      containerPort: port,
      healthCheck: {
        path: '/',
        timeoutSeconds: 4,
      },
    },
    containers: [
      {
        containerName: ecr.name,
        image: server.imageUri,
        ports: {
          [port]: 'HTTP',
        },
        environment: {
          PORT: port.toString(),
          ...containerEnv,
        },
      },
    ],
    serviceName: containerService.name,
  });

export const publicEndpoint = containerServiceDeployment.publicEndpoint;
export const deploymentState = containerServiceDeployment.state;
