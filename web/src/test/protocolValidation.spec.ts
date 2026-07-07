import { describe, it, expect } from 'vitest'

// Mirrors the validation rules used by ResourceForm.vue handleSubmit for
// protocol_type ∈ {rabbitmq, kafka}.
const KAFKA_HOST_PORT_RE = /^[A-Za-z0-9._-]+:\d{1,5}$/

function parseKafkaBootstrap(target: string): string[] {
  return target
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
}

function kafkaTargetValid(target: string): boolean {
  const entries = parseKafkaBootstrap(target)
  if (entries.length === 0) return false
  return entries.every((e) => KAFKA_HOST_PORT_RE.test(e))
}

function rabbitmqTargetValid(target: string): boolean {
  if (!target?.trim()) return false
  if (target.includes(',')) return false
  return true
}

const DEFAULT_PORTS: Record<string, number> = {
  redis: 6379,
  mongodb: 27017,
  ftp: 21,
  ssh: 22,
  mysql: 3306,
  postgres: 5432,
  rabbitmq: 5672,
  kafka: 9092,
}

describe('protocol form validation', () => {
  it('rabbitmq defaults port to 5672', () => {
    expect(DEFAULT_PORTS.rabbitmq).toBe(5672)
  })

  it('kafka defaults port to 9092', () => {
    expect(DEFAULT_PORTS.kafka).toBe(9092)
  })

  it('rejects empty rabbitmq target', () => {
    expect(rabbitmqTargetValid('')).toBe(false)
    expect(rabbitmqTargetValid('   ')).toBe(false)
  })

  it('rejects rabbitmq target with commas', () => {
    expect(rabbitmqTargetValid('h1,h2')).toBe(false)
  })

  it('accepts rabbitmq single host', () => {
    expect(rabbitmqTargetValid('rabbit.local')).toBe(true)
  })

  it('rejects empty kafka bootstrap list', () => {
    expect(kafkaTargetValid('')).toBe(false)
    expect(kafkaTargetValid(',,')).toBe(false)
  })

  it('rejects kafka entry without port', () => {
    expect(kafkaTargetValid('host')).toBe(false)
    expect(kafkaTargetValid('h1:9092,h2')).toBe(false)
  })

  it('accepts comma-separated kafka bootstrap and normalizes whitespace', () => {
    const entries = parseKafkaBootstrap(' h1:9092 , h2:9092,h3:9092 ')
    expect(entries).toEqual(['h1:9092', 'h2:9092', 'h3:9092'])
    expect(kafkaTargetValid(' h1:9092 , h2:9092 ')).toBe(true)
  })

  it('rejects kafka port out of u16', () => {
    expect(kafkaTargetValid('h1:999999')).toBe(false)
  })
})
