/**
 * EvalForge JavaScript/TypeScript SDK Client
 */

import axios, { AxiosInstance, AxiosError } from 'axios';
import { 
  TraceEvent, 
  TokenUsage, 
  EvalForgeConfig, 
  EventBatch,
  RetryOptions,
  Logger 
} from './types';
import { createLogger } from './utils/logger';
import { createRateLimiter } from './utils/rate-limiter';
import { generateId } from './utils/id-generator';

export class EvalForge {
  private config: Required<EvalForgeConfig>;
  private httpClient: AxiosInstance;
  private eventQueue: TraceEvent[] = [];
  private flushTimer: NodeJS.Timeout | null = null;
  private isDestroyed = false;
  private logger: Logger;
  private rateLimiter: any;

  constructor(config: EvalForgeConfig) {
    // Set defaults
    this.config = {
      apiKey: config.apiKey,
      projectId: config.projectId,
      baseUrl: config.baseUrl || 'https://api.evalforge.dev',
      batchSize: config.batchSize || 100,
      flushInterval: config.flushInterval || 5000,
      maxRetries: config.maxRetries || 3,
      timeout: config.timeout || 30000,
      debug: config.debug || false,
    };

    this.logger = createLogger(this.config.debug);
    this.rateLimiter = createRateLimiter(1000, 60000); // 1000 requests per minute

    // Initialize HTTP client
    this.httpClient = axios.create({
      baseURL: this.config.baseUrl,
      timeout: this.config.timeout,
      headers: {
        'Authorization': `Bearer ${this.config.apiKey}`,
        'Content-Type': 'application/json',
        'User-Agent': '@evalforge/sdk/0.1.0',
      },
    });

    // Start flush timer
    this.startFlushTimer();

    this.logger.debug('EvalForge client initialized', { 
      projectId: this.config.projectId,
      baseUrl: this.config.baseUrl,
      batchSize: this.config.batchSize,
      flushInterval: this.config.flushInterval,
    });
  }

  /**
   * Record a trace event
   */
  public trace(options: {
    operationType: string;
    input?: Record<string, any>;
    output?: Record<string, any>;
    metadata?: Record<string, any>;
    traceId?: string;
    spanId?: string;
    parentSpanId?: string;
    startTime?: Date;
    endTime?: Date;
    status?: 'success' | 'error' | 'timeout';
    tokens?: TokenUsage;
    cost?: number;
    provider?: string;
    model?: string;
  }): string {
    if (this.isDestroyed) {
      this.logger.warn('Cannot trace event: client is destroyed');
      return '';
    }

    const now = new Date();
    const traceId = options.traceId || generateId();
    const spanId = options.spanId || generateId();
    const startTime = options.startTime || now;
    const endTime = options.endTime || now;
    const durationMs = endTime.getTime() - startTime.getTime();

    const event: TraceEvent = {
      id: generateId(),
      project_id: this.config.projectId,
      trace_id: traceId,
      span_id: spanId,
      parent_span_id: options.parentSpanId,
      operation_type: options.operationType,
      start_time: startTime.toISOString(),
      end_time: endTime.toISOString(),
      duration_ms: Math.max(0, durationMs),
      status: options.status || 'success',
      input: options.input || {},
      output: options.output || {},
      metadata: options.metadata || {},
      tokens: options.tokens || { prompt: 0, completion: 0, total: 0 },
      cost: options.cost || 0,
      provider: options.provider || '',
      model: options.model || '',
    };

    this.enqueueEvent(event);
    return traceId;
  }

  /**
   * Manually flush all pending events
   */
  public async flush(timeoutMs: number = 10000): Promise<boolean> {
    if (this.isDestroyed) {
      return true;
    }

    this.logger.debug('Manually flushing events', { queueSize: this.eventQueue.length });

    if (this.eventQueue.length === 0) {
      return true;
    }

    const promise = this.flushEvents();
    const timeout = new Promise<boolean>((resolve) => {
      setTimeout(() => resolve(false), timeoutMs);
    });

    try {
      const result = await Promise.race([promise, timeout]);
      return result !== false;
    } catch (error) {
      this.logger.error('Error during flush', error);
      return false;
    }
  }

  /**
   * Close the client and flush all pending events
   */
  public async close(): Promise<void> {
    if (this.isDestroyed) {
      return;
    }

    this.logger.debug('Closing EvalForge client');
    this.isDestroyed = true;

    // Clear flush timer
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }

    // Flush remaining events
    await this.flush(5000);
    
    this.logger.debug('EvalForge client closed');
  }

  private enqueueEvent(event: TraceEvent): void {
    this.eventQueue.push(event);
    
    this.logger.debug('Event enqueued', { 
      eventId: event.id,
      queueSize: this.eventQueue.length,
      operationType: event.operation_type,
    });

    // Check if we should flush immediately
    if (this.eventQueue.length >= this.config.batchSize) {
      this.flushEvents().catch(error => {
        this.logger.error('Error flushing events from queue full condition', error);
      });
    }
  }

  private startFlushTimer(): void {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
    }

    this.flushTimer = setInterval(() => {
      if (this.eventQueue.length > 0) {
        this.flushEvents().catch(error => {
          this.logger.error('Error flushing events from timer', error);
        });
      }
    }, this.config.flushInterval);
  }

  private async flushEvents(): Promise<boolean> {
    if (this.eventQueue.length === 0) {
      return true;
    }

    const eventsToSend = this.eventQueue.splice(0, this.config.batchSize);
    
    this.logger.debug('Flushing events', { count: eventsToSend.length });

    try {
      await this.rateLimiter.checkLimit();
      await this.sendEvents(eventsToSend);
      
      this.logger.debug('Successfully sent events', { count: eventsToSend.length });
      return true;
    } catch (error) {
      this.logger.error('Failed to send events', error);
      
      // Re-add events to front of queue for retry
      this.eventQueue.unshift(...eventsToSend);
      return false;
    }
  }

  private async sendEvents(events: TraceEvent[]): Promise<void> {
    const batch: EventBatch = { events };
    const retryOptions: RetryOptions = {
      maxRetries: this.config.maxRetries,
      baseDelay: 1000,
      maxDelay: 10000,
      jitter: true,
    };

    return this.retryRequest(() => 
      this.httpClient.post('/api/events', batch), 
      retryOptions
    );
  }

  private async retryRequest<T>(
    request: () => Promise<T>, 
    options: RetryOptions
  ): Promise<T> {
    let lastError: Error;

    for (let attempt = 0; attempt <= options.maxRetries; attempt++) {
      try {
        const result = await request();
        return result;
      } catch (error) {
        lastError = error as Error;

        if (attempt === options.maxRetries) {
          break;
        }

        const isRetryable = this.isRetryableError(error);
        if (!isRetryable) {
          break;
        }

        const delay = this.calculateDelay(attempt, options);
        this.logger.debug(`Request failed, retrying in ${delay}ms`, { 
          attempt: attempt + 1,
          maxRetries: options.maxRetries,
          error: error instanceof Error ? error.message : 'Unknown error',
        });

        await this.sleep(delay);
      }
    }

    throw lastError!;
  }

  private isRetryableError(error: any): boolean {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      
      // Retry on network errors
      if (!axiosError.response) {
        return true;
      }

      // Retry on server errors and rate limits
      const status = axiosError.response.status;
      return status >= 500 || status === 429;
    }

    return false;
  }

  private calculateDelay(attempt: number, options: RetryOptions): number {
    const exponentialDelay = Math.min(
      options.baseDelay * Math.pow(2, attempt),
      options.maxDelay
    );

    if (options.jitter) {
      // Add ±25% jitter
      const jitter = exponentialDelay * 0.25 * (Math.random() - 0.5);
      return Math.max(0, exponentialDelay + jitter);
    }

    return exponentialDelay;
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// Global client instance
let globalClient: EvalForge | null = null;

/**
 * Configure the global EvalForge client
 */
export function configure(config: EvalForgeConfig): EvalForge {
  if (globalClient) {
    globalClient.close();
  }
  
  globalClient = new EvalForge(config);
  return globalClient;
}

/**
 * Get the global EvalForge client
 */
export function getClient(): EvalForge | null {
  return globalClient;
}

/**
 * Record a trace event using the global client
 */
export function trace(options: Parameters<EvalForge['trace']>[0]): string {
  if (!globalClient) {
    throw new Error('EvalForge client not configured. Call configure() first.');
  }
  
  return globalClient.trace(options);
}

/**
 * Flush all pending events using the global client
 */
export async function flush(timeoutMs?: number): Promise<boolean> {
  if (!globalClient) {
    return true;
  }
  
  return globalClient.flush(timeoutMs);
}