library ieee;
use ieee.std_logic_1164.all;
use ieee.numeric_std.all;

entity alu is
    port (
        a, b : in  std_logic_vector(7 downto 0);
        op   : in  std_logic_vector(1 downto 0);
        result : out std_logic_vector(7 downto 0)
    );
end entity alu;

architecture behavioral of alu is
    signal temp : std_logic_vector(7 downto 0);
    constant ZERO : std_logic_vector(7 downto 0) := (others => '0');
    type state_type is (idle, running, done);
    component adder
        port (
            x, y : in  std_logic_vector(7 downto 0);
            sum  : out std_logic_vector(7 downto 0)
        );
    end component adder;
begin
    compute: process(a, b, op)
    begin
        case op is
            when "00" => temp <= std_logic_vector(unsigned(a) + unsigned(b));
            when "01" => temp <= std_logic_vector(unsigned(a) - unsigned(b));
            when others => temp <= ZERO;
        end case;
    end process compute;

    result <= temp;
end architecture behavioral;

package alu_pkg is
    function max_val(width : integer) return integer;
    procedure reset_reg(signal reg : out std_logic_vector);
end package alu_pkg;
