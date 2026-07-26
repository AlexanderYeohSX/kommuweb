#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { PaymentStack } from '../lib/payment-stack';

const app = new cdk.App();
new PaymentStack(app, 'KommuPaymentStack', {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION || 'ap-southeast-2',
  },
});
