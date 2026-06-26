import { EventEmitter } from 'events';

export class ResilientStream<T> extends EventEmitter {
    private buffer: T[] = [];
    private readonly maxBufferSize = 1000;

    push(item: T) {
        this.buffer.push(item);
        if (this.buffer.length > this.maxBufferSize) {
            this.buffer.shift();
        }
        this.emit('data', item);
    }

    async *[Symbol.asyncIterator]() {
        // First yield everything in the buffer
        for (const item of this.buffer) {
            yield item;
        }

        // Then yield new items as they arrive
        while (true) {
            yield await new Promise<T>((resolve) => {
                this.once('data', resolve);
            });
        }
    }

    clear() {
        this.buffer = [];
    }
}
