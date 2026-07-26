import * as apigwv2 from 'aws-cdk-lib/aws-apigatewayv2';
import * as cdk from 'aws-cdk-lib';
import * as cr from 'aws-cdk-lib/custom-resources';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as s3assets from 'aws-cdk-lib/aws-s3-assets';
import { Construct } from 'constructs';
import { execSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';

/** Discovered from AWS account (ap-southeast-2). */
const CURLEC_FUNCTION_NAME = 'CurlecGateway';
const CURLEC_FUNCTION_ARN =
  'arn:aws:lambda:ap-southeast-2:730335489833:function:CurlecGateway';
const HTTP_API_ID = 'ifhdr5efvk';
const HTTP_API_ENDPOINT =
  'https://ifhdr5efvk.execute-api.ap-southeast-2.amazonaws.com';
const CUSTOM_DOMAIN = 'aws.kommu.ai';
const PAYMENT_INTEGRATION_ID = 'ghwjun3';
const STANDARD_CHECKOUT_ROUTES = [
  'POST /curlec/orders',
  'POST /curlec/subscriptions',
  'POST /curlec/verify',
  'POST /curlec/webhook',
];

export class PaymentStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const paymentDir = path.join(__dirname, '../../payment');

    // Provisioned 1/1 stays inside DynamoDB always-free tier at Kommu volume.
    const checkoutContext = new dynamodb.Table(this, 'CheckoutContext', {
      tableName: 'kommu-checkout-context',
      partitionKey: { name: 'entity_id', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PROVISIONED,
      readCapacity: 1,
      writeCapacity: 1,
      timeToLiveAttribute: 'ttl',
      removalPolicy: cdk.RemovalPolicy.RETAIN,
    });

    const paymentFn = lambda.Function.fromFunctionAttributes(this, 'CurlecGatewayFn', {
      functionArn: CURLEC_FUNCTION_ARN,
      sameEnvironment: true,
    });

    checkoutContext.grantReadWriteData(paymentFn);

    const codeAsset = new s3assets.Asset(this, 'PaymentLambdaCode', {
      path: paymentDir,
      bundling: {
        image: cdk.DockerImage.fromRegistry('alpine'),
        command: ['echo', 'unused'],
        local: {
          tryBundle(outputDir: string): boolean {
            execSync('make zip', { cwd: paymentDir, stdio: 'inherit' });
            const zipPath = path.join(paymentDir, 'myFunction.zip');
            fs.copyFileSync(zipPath, path.join(outputDir, 'code.zip'));
            return true;
          },
        },
        outputType: cdk.BundlingOutput.ARCHIVED,
      },
    });

    const updateCode = new cr.AwsCustomResource(this, 'UpdateCurlecGatewayCode', {
      installLatestAwsSdk: false,
      onCreate: {
        service: 'Lambda',
        action: 'updateFunctionCode',
        parameters: {
          FunctionName: CURLEC_FUNCTION_NAME,
          S3Bucket: codeAsset.s3BucketName,
          S3Key: codeAsset.s3ObjectKey,
        },
        physicalResourceId: cr.PhysicalResourceId.of(codeAsset.assetHash),
      },
      onUpdate: {
        service: 'Lambda',
        action: 'updateFunctionCode',
        parameters: {
          FunctionName: CURLEC_FUNCTION_NAME,
          S3Bucket: codeAsset.s3BucketName,
          S3Key: codeAsset.s3ObjectKey,
        },
        physicalResourceId: cr.PhysicalResourceId.of(codeAsset.assetHash),
      },
      policy: cr.AwsCustomResourcePolicy.fromStatements([
        new iam.PolicyStatement({
          actions: ['lambda:UpdateFunctionCode'],
          resources: [CURLEC_FUNCTION_ARN],
        }),
        new iam.PolicyStatement({
          actions: ['s3:GetObject'],
          resources: [codeAsset.bucket.arnForObjects('*')],
        }),
      ]),
    });
    updateCode.node.addDependency(codeAsset);

    STANDARD_CHECKOUT_ROUTES.forEach((routeKey, index) => {
      new apigwv2.CfnRoute(this, `StandardCheckoutRoute${index}`, {
        apiId: HTTP_API_ID,
        routeKey,
        target: `integrations/${PAYMENT_INTEGRATION_ID}`,
      });
      const routePath = routeKey.replace(/^POST /, '');
      paymentFn.addPermission(`ApiInvoke${index}`, {
        principal: new iam.ServicePrincipal('apigateway.amazonaws.com'),
        sourceArn: `arn:aws:execute-api:${cdk.Stack.of(this).region}:${cdk.Stack.of(this).account}:${HTTP_API_ID}/*/*${routePath}`,
      });
    });

    new cdk.CfnOutput(this, 'LambdaFunctionName', { value: CURLEC_FUNCTION_NAME });
    new cdk.CfnOutput(this, 'HttpApiId', { value: HTTP_API_ID });
    new cdk.CfnOutput(this, 'HttpApiEndpoint', { value: HTTP_API_ENDPOINT });
    new cdk.CfnOutput(this, 'CustomDomain', { value: CUSTOM_DOMAIN });
    new cdk.CfnOutput(this, 'CheckoutContextTable', { value: checkoutContext.tableName });
    new cdk.CfnOutput(this, 'SetEnvHint', {
      value:
        'aws lambda update-function-configuration --function-name CurlecGateway --region ap-southeast-2 ' +
        '--environment "Variables={...,CHECKOUT_CONTEXT_TABLE=kommu-checkout-context,CURLEC_DEV_TEST_PLAN_ID=plan_xxx}" ' +
        '(merge with existing vars in console — do not replace blindly)',
    });
  }
}
