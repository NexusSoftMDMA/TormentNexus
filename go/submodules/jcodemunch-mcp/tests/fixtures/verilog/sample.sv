`include "defines.vh"
`include "utils.svh"
`define DATA_WIDTH 8

module alu #(
    parameter WIDTH = 8
)(
    input  logic [WIDTH-1:0] a,
    input  logic [WIDTH-1:0] b,
    input  logic [1:0]       op,
    output logic [WIDTH-1:0] result
);
    localparam ADD = 2'b00;
    localparam SUB = 2'b01;

    function logic [WIDTH-1:0] compute;
        input logic [WIDTH-1:0] x, y;
        input logic [1:0] operation;
        case (operation)
            ADD: compute = x + y;
            SUB: compute = x - y;
            default: compute = '0;
        endcase
    endfunction

    task display_result;
        $display("Result: %h", result);
    endtask

    assign result = compute(a, b, op);
endmodule

interface axi_if #(
    parameter ADDR_WIDTH = 32
);
    logic [ADDR_WIDTH-1:0] addr;
    logic valid;
    logic ready;
endinterface

package alu_pkg;
    typedef enum logic [1:0] {
        OP_ADD = 2'b00,
        OP_SUB = 2'b01
    } op_t;
endpackage

class scoreboard;
    function void check(input logic [7:0] expected, actual);
        if (expected !== actual)
            $error("Mismatch: %h != %h", expected, actual);
    endfunction
endclass
