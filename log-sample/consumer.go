package main

import (
	"fmt"
	"sync"

	"github.com/Shopify/sarama"
)

func consumer() {
	fmt.Println("consumer begin...")
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	wg := sync.WaitGroup{}
	// 创建消费者
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		fmt.Println("consumer created failed, error is ", err.Error())
		return
	}
	defer consumer.Close()

	// Partitions(topic): 返回该topic的所有分区id
	partitionList, err := consumer.Partitions("test")
	if err != nil {
		fmt.Println("get consumer failed :", err.Error())
		return
	}

	for partition := range partitionList {
		// ConsumePartition()方法根据topic分区和给定的偏移量创建相应的分区消费者
		// 如果该消费者已经消费了该信息会返回error， OffsetNewest 消费最新的消息
		pc, err := consumer.ConsumePartition("test", int32(partition), sarama.OffsetNewest)
		if err != nil {
			panic(err)
		}
		// 异步关闭，保证数据落盘
		defer pc.AsyncClose()
		wg.Add(1)
		go func(sarama.PartitionConsumer) {
			defer wg.Done()
			// Message() 方法返回一个消费消息类型的只读通道，由代理产生
			for msg := range pc.Messages() {
				fmt.Printf("%s---Partition:%d, Offset:%d, Key:%s, Value:%s\n",
					msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
			}
		}(pc)
	}
	wg.Wait()
	consumer.Close()
}
