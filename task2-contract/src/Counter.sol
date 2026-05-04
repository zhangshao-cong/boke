// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title Counter
 * @dev 一个简单的计数器合约，用于演示 abigen 代码生成
 * 包含基本的递增、递减和查询功能
 */
contract Counter {
    // 状态变量：存储当前计数器的值
    uint256 private count;

    // 事件：计数器值改变时触发
    event CountChanged(uint256 oldValue, uint256 newValue);

    // 事件：计数器被重置时触发
    event CountReset(address indexed resetBy);

    /**
     * @dev 构造函数，初始化计数器为 0
     */
    constructor() {
        count = 0;
    }

    /**
     * @dev 获取当前计数器的值
     * @return 当前计数值
     */
    function getCount() public view returns (uint256) {
        return count;
    }

    /**
     * @dev 递增计数器
     * 每次调用 count 增加 1
     */
    function increment() public {
        uint256 oldValue = count;
        count += 1;
        emit CountChanged(oldValue, count);
    }

    /**
     * @dev 递减计数器
     * 每次调用 count 减少 1，不能低于 0
     */
    function decrement() public {
        require(count > 0, "Counter: cannot decrement below zero");
        uint256 oldValue = count;
        count -= 1;
        emit CountChanged(oldValue, count);
    }

    /**
     * @dev 增加指定数值
     * @param value 要增加的数值
     */
    function add(uint256 value) public {
        uint256 oldValue = count;
        count += value;
        emit CountChanged(oldValue, count);
    }

    /**
     * @dev 设置计数器为指定值
     * @param value 新的计数值
     */
    function setCount(uint256 value) public {
        uint256 oldValue = count;
        count = value;
        emit CountChanged(oldValue, count);
    }

    /**
     * @dev 重置计数器为 0
     */
    function reset() public {
        uint256 oldValue = count;
        count = 0;
        emit CountChanged(oldValue, 0);
        emit CountReset(msg.sender);
    }
}
