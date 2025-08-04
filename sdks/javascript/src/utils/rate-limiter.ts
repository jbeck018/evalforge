/**
 * Rate limiting utilities
 */

export interface RateLimiter {
  checkLimit(): Promise<void>;
  reset(): void;
}

class TokenBucketRateLimiter implements RateLimiter {
  private tokens: number;
  private lastRefill: number;

  constructor(
    private maxTokens: number,
    private refillPeriodMs: number
  ) {
    this.tokens = maxTokens;
    this.lastRefill = Date.now();
  }

  async checkLimit(): Promise<void> {
    this.refillTokens();

    if (this.tokens >= 1) {
      this.tokens -= 1;
      return;
    }

    // Calculate wait time
    const timeToWait = this.refillPeriodMs / this.maxTokens;
    await this.sleep(timeToWait);
    
    // Recursively check again
    return this.checkLimit();
  }

  reset(): void {
    this.tokens = this.maxTokens;
    this.lastRefill = Date.now();
  }

  private refillTokens(): void {
    const now = Date.now();
    const timePassed = now - this.lastRefill;
    const tokensToAdd = (timePassed / this.refillPeriodMs) * this.maxTokens;
    
    this.tokens = Math.min(this.maxTokens, this.tokens + tokensToAdd);
    this.lastRefill = now;
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

export function createRateLimiter(
  maxTokens: number, 
  refillPeriodMs: number
): RateLimiter {
  return new TokenBucketRateLimiter(maxTokens, refillPeriodMs);
}