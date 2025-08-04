/**
 * Decorators and context managers for tracing
 */

import { getClient } from './client';
import { generateId } from './utils/id-generator';
import { TokenUsage } from './types';

/**
 * Decorator for automatic function tracing
 */
export function trace(options: {
  operationType?: string;
  captureArgs?: boolean;
  captureResult?: boolean;
  captureExceptions?: boolean;
  provider?: string;
  model?: string;
} = {}) {
  return function<T extends Function>(target: T): T {
    const originalFunction = target as any;
    
    const wrapper = function(this: any, ...args: any[]) {
      const client = getClient();
      if (!client) {
        return originalFunction.apply(this, args);
      }

      const traceId = generateId();
      const spanId = generateId();
      const startTime = new Date();
      const operationType = options.operationType || originalFunction.name || 'function_call';

      // Capture arguments
      const inputData: any = {};
      if (options.captureArgs !== false && args.length > 0) {
        try {
          inputData.arguments = args.map(arg => {
            if (typeof arg === 'object' && arg !== null) {
              return JSON.parse(JSON.stringify(arg)); // Deep clone
            }
            return arg;
          });
        } catch (error) {
          inputData.arguments = '<serialization_error>';
        }
      }

      try {
        const result = originalFunction.apply(this, args);

        // Handle promises
        if (result && typeof result.then === 'function') {
          return result
            .then((resolvedResult: any) => {
              const endTime = new Date();
              
              const outputData: any = {};
              if (options.captureResult !== false) {
                try {
                  outputData.result = typeof resolvedResult === 'object' 
                    ? JSON.parse(JSON.stringify(resolvedResult))
                    : resolvedResult;
                } catch (error) {
                  outputData.result = '<serialization_error>';
                }
              }

              client.trace({
                operationType,
                input: inputData,
                output: outputData,
                metadata: {
                  function_name: originalFunction.name,
                  is_async: true,
                },
                traceId,
                spanId,
                startTime,
                endTime,
                status: 'success',
                provider: options.provider,
                model: options.model,
              });

              return resolvedResult;
            })
            .catch((error: any) => {
              const endTime = new Date();
              
              const outputData: any = {};
              if (options.captureExceptions !== false) {
                outputData.error = error instanceof Error ? error.message : 'Unknown error';
                outputData.error_type = error instanceof Error ? error.constructor.name : 'UnknownError';
              }

              client.trace({
                operationType,
                input: inputData,
                output: outputData,
                metadata: {
                  function_name: originalFunction.name,
                  is_async: true,
                },
                traceId,
                spanId,
                startTime,
                endTime,
                status: 'error',
                provider: options.provider,
                model: options.model,
              });

              throw error;
            });
        }

        // Handle synchronous functions
        const endTime = new Date();
        
        const outputData: any = {};
        if (options.captureResult !== false) {
          try {
            outputData.result = typeof result === 'object' && result !== null
              ? JSON.parse(JSON.stringify(result))
              : result;
          } catch (error) {
            outputData.result = '<serialization_error>';
          }
        }

        client.trace({
          operationType,
          input: inputData,
          output: outputData,
          metadata: {
            function_name: originalFunction.name,
            is_async: false,
          },
          traceId,
          spanId,
          startTime,
          endTime,
          status: 'success',
          provider: options.provider,
          model: options.model,
        });

        return result;
      } catch (error) {
        const endTime = new Date();
        
        const outputData: any = {};
        if (options.captureExceptions !== false) {
          outputData.error = error instanceof Error ? error.message : 'Unknown error';
          outputData.error_type = error instanceof Error ? error.constructor.name : 'UnknownError';
        }

        client.trace({
          operationType,
          input: inputData,
          output: outputData,
          metadata: {
            function_name: originalFunction.name,
            is_async: false,
          },
          traceId,
          spanId,
          startTime,
          endTime,
          status: 'error',
          provider: options.provider,
          model: options.model,
        });

        throw error;
      }
    };

    // Preserve original function properties
    Object.defineProperty(wrapper, 'name', { value: originalFunction.name });
    Object.defineProperty(wrapper, 'length', { value: originalFunction.length });

    return wrapper as T;
  };
}

/**
 * Context manager for creating nested traces
 */
export class TraceContext {
  public readonly traceId: string;
  public readonly spanId: string;
  private startTime: Date;
  private endTime?: Date;
  private outputData: Record<string, any> = {};

  constructor(
    private operationType: string,
    private inputData: Record<string, any> = {},
    private metadata: Record<string, any> = {},
    private provider?: string,
    private model?: string,
    private parentSpanId?: string
  ) {
    this.traceId = generateId();
    this.spanId = generateId();
    this.startTime = new Date();
  }

  /**
   * Create a child context
   */
  child(
    operationType: string,
    inputData: Record<string, any> = {},
    metadata: Record<string, any> = {},
    provider?: string,
    model?: string
  ): TraceContext {
    return new TraceContext(
      operationType,
      inputData,
      metadata,
      provider || this.provider,
      model || this.model,
      this.spanId
    );
  }

  /**
   * Set output data for this trace
   */
  setOutput(data: Record<string, any>): void {
    this.outputData = { ...this.outputData, ...data };
  }

  /**
   * Add metadata to this trace
   */
  addMetadata(key: string, value: any): void {
    this.metadata[key] = value;
  }

  /**
   * Manually finish the trace
   */
  finish(status: 'success' | 'error' | 'timeout' = 'success'): void {
    if (this.endTime) {
      return; // Already finished
    }

    const client = getClient();
    if (!client) {
      return;
    }

    this.endTime = new Date();

    client.trace({
      operationType: this.operationType,
      input: this.inputData,
      output: this.outputData,
      metadata: this.metadata,
      traceId: this.traceId,
      spanId: this.spanId,
      parentSpanId: this.parentSpanId,
      startTime: this.startTime,
      endTime: this.endTime,
      status,
      provider: this.provider,
      model: this.model,
    });
  }

  /**
   * Execute a function within this trace context
   */
  async run<T>(fn: () => T | Promise<T>): Promise<T> {
    try {
      const result = await fn();
      this.finish('success');
      return result;
    } catch (error) {
      if (error instanceof Error) {
        this.setOutput({
          error: error.message,
          error_type: error.constructor.name,
        });
      }
      this.finish('error');
      throw error;
    }
  }
}